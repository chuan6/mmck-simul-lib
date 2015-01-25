// Package mmx implements simulation of M/M/c/k queueing systems.
// Usage:
//
//  simtask := mmx.NewEnvironment()
//  simtask.Arrive(_arrival_rate_)
//  simtask.Line  (_line_capacity_)
//  simtask.Serve (_number_of_servers_, _service_rate_per_server_)
//
//  rejected, departed := simtask.Output()
//  var c mmx.Customer
//  for i := 0; i < _ncustomers_; i++ {
//      select {
//      case c = <-departed:
//          // ask the departed customer c
// 	case c = <-rejected:
//          // ask the rejected customer c
//      }
//  }

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

// An Environment is a set of resources local to one simulation.
type Environment struct {
	acc chan float64  // stream of arrival times of accepted customers
	rej chan Customer // stream of rejected customers
	srv chan float64  // stream of start servicing times of customers in line
	dep chan Customer // stream of departed customers
	chl chan float64  // stream of time points when line is available
	chs chan float64  // stream of time points when server is available
	cus Customer      // the current customer
}

// Create one Environment for one simulation.
func NewEnvironment() (e *Environment) {
	e = new(Environment)
	e.acc = make(chan float64)
	e.rej = make(chan Customer)
	e.srv = make(chan float64)
	e.dep = make(chan Customer)
	e.chl = make(chan float64)
	e.chs = make(chan float64)
	return
}

func (e *Environment) Output() (rej, dep <-chan Customer) {
	rej = e.rej
	dep = e.dep
	return
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
func (e *Environment) Arrive(rate float64) {
	var now float64
	gen := newExpRNG(rate)
	go func() {
		// accept the 1st customer
		e.acc <- 0.0
		for {
			t := <-e.chl // line will have space at t
			now += gen()
			for now < t {
				e.rej <- Customer{T0: now}
				now += gen()
			} // now >= t
			e.cus.T0 = now // set arrival time of the current customer
			e.acc <- now   // notice Line the new accepted arrival
		}
	}()
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
func (e *Environment) Line(k int) {
	q := makeline(k)

	go func() {
		// handle the 1st accepted customer
		t0 := <-e.acc
		q.arr[q.i] = t0
		e.cus.T1 = t0
		e.srv <- t0
		e.chl <- t0

		var t1 float64
		for {
			t0 = <-e.acc
			t1 = <-e.chs
			if t0 >= t1 {
				t1 = t0
				q.arr[q.i] = t1
			} else {
				q.arr[q.i] = t1
				q.i = q.isuc()
			}
			e.cus.T1 = t1
			e.srv <- t1
			e.chl <- q.arr[q.i]
		}
	}()
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
func (e *Environment) Serve(c int, rate float64) {
	h := makegroup(c, rate)
	go func() {
		// depart the 1st customer
		t1 := <-e.srv
		e.cus.T2, e.cus.SrvrID = h.gen(t1)
		e.dep <- e.cus
		e.chs <- h[0].now // the current time of the server on top of the heap
		for {
			t1 = <-e.srv
			e.cus.T2, e.cus.SrvrID = h.gen(t1)
			e.dep <- e.cus
			e.chs <- h[0].now
		}
	}()
}
