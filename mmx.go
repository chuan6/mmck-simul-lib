/* Package mmx implements simulation of M/M/c/k queueing systems. */

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

type Arrival interface {
	Arrive() float64
}

type Line interface {
	WaitOrPass(float64, float64) float64
}

type Service interface {
	Serve(float64) (float64, int)
}

type ExpArrival struct {
	src  *rand.Rand
	rate float64
}

func NewExpArrival(rate float64) (e *ExpArrival) {
	e = new(ExpArrival)
	e.src = rand.New(rand.NewSource(time.Now().UnixNano()))
	e.rate = rate
	return
}

func (e *ExpArrival) Arrive() float64 {
	return e.src.ExpFloat64() / e.rate
}

type FifoLine struct {
	arr  []float64
	back int
}

func NewFifoLine(c int) (q FifoLine) {
	q.arr = make([]float64, c)
	return
}

func (q *FifoLine) isuc() int {
	return (q.back + 1) % cap(q.arr)
}

func (q *FifoLine) WaitOrPass(t0, t float64) (t1, chl float64) {
	if t0 < t {// wait
		q.arr[q.back] = t
		q.back = q.isuc()
	} else {// pass
		t = t0
		q.arr[q.back] = t
	}
	t1 = t
	chl = q.arr[q.back]
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

// A server is a unit resource of the server group.
type server struct {
	id int
	now float64
	gen ExpRNG
}

// A MinheapExpService is servers arranged in min-heap (according to their clock).
type MinheapExpService [](*server) // where heap is built upon

func (h MinheapExpService) min(i, j int) int {
	if h[j].now < h[i].now {
		return j
	}
	return i
}

// Given the i-th element within the heap, return the index of the
// minimum of the i-th element, its left child, and its right child.
func (h MinheapExpService) minOfTri(i int) int {
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
func (h MinheapExpService) Serve(now float64) (depTime float64, sid int) {
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

// Make a MinheapExpService of n servers, all of which are specified having service rate r.
func makeMinheapExpService(n int, r float64) (h MinheapExpService) {
	h = make([]*server, n)
	p := make([]server, n) // pointer to the underlying array that stores servers
	for i := 0; i < n; i++ {
		p[i].id = i
		p[i].gen = newExpRNG(r)
		h[i] = &p[i]
	}
	return
}

func Combined(arate float64, narr int, c int, srate float64, n int) (rejs, deps <-chan Customer) {
	/* all parameters should be none negative */
	
	var chl, chs float64
	rej := make(chan Customer, 8)
	dep := make(chan Customer, 32)

	arrgen := newExpRNG(arate)
	q := Newline(c)
	h := makeMinheapExpService(n, srate)

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
			if t0 >= t1 {// be served immediately
				t1 = t0
				q.arr[q.i] = t1
			} else {// wait til available
				q.arr[q.i] = t1
				q.i = q.isuc()
			}
			chl = q.arr[q.i] // update chl
			// waited

			t2, sid = h.gen(t1)
			dep <- Customer{T0: t0, T1: t1, T2: t2, SrvrID: sid}
			chs = h[0].now // update chs
			// served and departed
		}
	}()

	rejs = rej
	deps = dep
	return
}
