/* Package mmx implements simulation of M/M/c/k queueing systems.
 Usage:

Serve(Line(Arrive(_arrival_rate_), _line_capacity_), _nservers_, _service_rate_per_server)

  simtask := mmx.NewEnvironment()
  simtask.Arrive(_arrival_rate_)
  simtask.Line  (_line_capacity_)
  simtask.Serve (_number_of_servers_, _service_rate_per_server_)

  rejected, departed := simtask.Output()
  var c mmx.Customer
  for i := 0; i < _ncustomers_; i++ {
      select {
      case c = <-departed:
          // ask the departed customer c
 	case c = <-rejected:
          // ask the rejected customer c
      }
  }
*/

package mmx

import (
	"math/rand"
	"time"
)

type Customer struct {
	T0     float64 // time of arrival of the customer
	T1     float64 // time of service of the customer
	T2     float64 // time of departure of the customer
	SrvrID int     // identifier of the server that serves the customer
}

type Inter struct {
	recv <-chan Customer
	send chan<- float64
}

// The exponential random number generator type, whose value, a function,
// returns a "real" random number every time it is called. And the sequence
// of number generated through a number of calls follow exponential
// distribution.
type ExpRNG func() float64

// Initialize a new ExpRNG according to the given rate.
func newExpRNG(rate float64) ExpRNG {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func() float64 {
		return r.ExpFloat64() / rate
	}
}

// Go generate arrivals!
func Arrive(rate float64) (io Inter, rejected <-chan Customer) {
	gen := newExpRNG(rate)
	acc := make(chan Customer)
	chl := make(chan float64)
	rej := make(chan Customer)

	go func() {
		var t0, now float64

		// accept the 1st customer
		acc <- Customer{T0: 0.0}
		for {
			now = <-chl // line have space at now
			t0 += gen()
			for t0 < now {
				rej <- Customer{T0: t0}
				t0 += gen()
			} // t0 >= now
			acc <- Customer{T0: t0}
		}
	}()
	
	io.recv = acc
	io.send = chl
	rejected = rej
	return
}

// A line is waiting positions arranged in queue.
// Note: "queue" is referred here as a FIFO container with fixed capacity. Its
// current implementation resembles ring buffer.
type line struct {
	arr []float64 // refers to an underlying array that implements ring buffer
	i   int    // index to the back of the queue (inclusive)
}

func makeline(n int) (q line) {
	q.arr = make([]float64, n)
	return
}

func (q *line) isuc() int {
	return (q.i + 1) % cap(q.arr)
}

// Go manage the FIFO waiting line!
func Line(io Inter, c int) Inter {
	q := makeline(c)
	acc := io.recv
	chl := io.send
	srv := make(chan Customer)
	chs := make(chan float64)

	go func() {
		cus := <-acc
		q.arr[q.i] = cus.T0
		cus.T1 = cus.T0
		srv <- cus
		chl <- cus.T0

		var now float64
		for {
			cus = <-acc
			now = <-chs
			if cus.T0 >= now {
				now = cus.T0
				q.arr[q.i] = now
			} else {
				q.arr[q.i] = now
				q.i = q.isuc()
			}
			cus.T1 = now
			srv <- cus
			chl <- q.arr[q.i]
		}
	}()
	
	return Inter{recv: srv, send: chs}
}

// A server is a unit resource of the server group.
type server struct {
	id int
	now float64
	gen ExpRNG
}

// A group is servers arranged in min-heap (according to their clock).
type group [](*server) // where heap is built upon

func (h group) min(i, j int) int {
	if h[j].now < h[i].now {
		return j
	}
	return i
}

// Given the i-th element within the heap, return the index of the
// minimum of the i-th element, its left child, and its right child.
func (h group) minOfTri(i int) int {
	j := 2*i + 1
	k := j + 1
	limit := len(h) - 1

	switch {
	case j > limit:
		return i
	case k > limit:
		return h.min(i, j)
	case h.min(i, j) == i:
		return h.min(i, k)
	default:
		return h.min(j, k)
	}
}

// Given start servicing time, return a scheduled departure time,
// and the corresponding server ID. Properties of the heap is maintained.
func (h group) gen(now float64) (depTime float64, sid int) {
	sid = h[0].id
	depTime = now + h[0].gen()
	h[0].now = depTime

	// maintain the heap
	s := 0
	for t := h.minOfTri(s); s != t; t = h.minOfTri(s) {
		h[s], h[t] = h[t], h[s] // floating down
		s = t
	}
	return
}

// Make a group of n servers, all of which are specified having service rate r.
func makegroup(n int, r float64) (h group) {
	h = make([]*server, n)
	p := make([]server, n) // pointer to the underlying array that stores servers
	for i := 0; i < n; i++ {
		p[i].id = i
		p[i].gen = newExpRNG(r)
		h[i] = &p[i]
	}
	return
}

// Go schedule departures for incomming customers!
func Serve(io Inter, n int, rate float64) (departed <-chan Customer) {
	h := makegroup(n, rate)
	srv := io.recv
	chs := io.send
	dep := make(chan Customer)

	go func() {
		// depart the 1st customer
		cus := <-srv
		cus.T2, cus.SrvrID = h.gen(cus.T1)
		dep <- cus
		chs <- h[0].now // the current time of the server on top of the heap

		for {
			cus = <-srv
			cus.T2, cus.SrvrID = h.gen(cus.T1)
			dep <- cus
			chs <- h[0].now
		}
	}()
	
	departed = dep
	return
}

func Combined(arate float64, narr int, c int, srate float64, n int) (rejs, deps <-chan Customer) {
	/* all parameters should be none negative */
	
	var chl, chs float64
	rej := make(chan Customer)
	dep := make(chan Customer)

	arrgen := newExpRNG(arate)
	q := makeline(c)
	h := makegroup(n, srate)

	go func() {
		dep <- Customer{} // first departure

		var t0, t1, t2 float64
		var sid int
		for i := 0; i < narr; i++ {
			t0 += arrgen()
			for t0 < chl {
				rej <- Customer{T0: t0}
				t0 += arrgen()
			} // t0 >= chl
			// accepted, or rejected

			t1 = chs
			if t0 >= t1 {
				t1 = t0
				q.arr[q.i] = t1
			} else {
				q.arr[q.i] = t1
				q.i = q.isuc()
			}
			chl = q.arr[q.i] // update chl
			// waited

			t2, sid = h.gen(t1)
			dep <- Customer{T0: t0, T1: t1, T2: t2, SrvrID: sid}
			chs = h[0].now
			// served and departed
		}
	}()

	rejs = rej
	deps = dep
	return
}
