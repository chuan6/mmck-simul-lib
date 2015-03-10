package mmck

import (
	"math/rand"
	"time"
)

// This default implementation produces the same results as analytical approach
// from tools such as Mathematica.

// exponential interval of arrivals; implements Arrival interface
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

func (e *ExpArrival) ArriveIn() float64 {
	return e.src.ExpFloat64() / e.rate
}

// fixed length queue that implements Line interface
type Ring struct {
	arr  []float64
	back int
}

func NewRing(c int) (q Ring) {
	q.arr = make([]float64, c)
	return
}

func (q Ring) WaitOrPass(t0, t float64) (t1 float64, sid int) {
	if t0 < t { // wait
		q.arr[q.back] = t
	} else { // pass
		q.arr[q.back] = t0
	}
	t1 = q.arr[q.back]
	sid = q.back

	q.back = (q.back + 1) % cap(q.arr)
	return
}

// Note: the FIFO property is imposed by the minheap service module instead.
func (q Ring) Next() float64 {
	return q.arr[q.back]
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
	id  int
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

// Guarantees that the eariliest available time of the service module is
// returned.
func (h MinheapExpService) Next() float64 {
	return h[0].now
}

// Make a MinheapExpService of n servers, all of which are specified having service rate r.
func MakeMinheapExpService(n int, r float64) (h MinheapExpService) {
	h = make([]*server, n)
	p := make([]server, n) // pointer to the underlying array that stores servers
	for i := 0; i < n; i++ {
		p[i].id = i
		p[i].gen = newExpRNG(r)
		h[i] = &p[i]
	}
	return
}
