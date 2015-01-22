// Package mmx implements simulation of M/M/c/k queueing systems.
// Usage:
// 	simtask := mmx.NewEnvironment()
//	mmx.Arrive(simtask, _arrival_rate_)
//	mmx.Line  (simtask, _line_capacity_)
//	mmx.Serve (simtask, _number_of_servers_, _service_rate_per_server_)
//
//	// statistical analysis on output from the rejection channel, and
//	// departure channel
//	var c mmx.Customer
//	for i := 0; i < _narrivals_; i++ {
//		select {
//		case c = <-simtask.Dep:
//			// handle the departed customer c
//		caes c = <-simtask.Rej:
//			// handle the rejected customer c
//		}
//	}
//	...
package mmx

import (
	"math/rand"
	"time"
)

type Customer struct {
	T0     float64 // time of arrival of the customer
	T1     float64 // time of service of the customer
	T2     float64 // time of departure of the customer
	SeatID int     // identifier of the seat that seats the customer
	SrvrID int     // identifier of the server that serves the customer
}

// An Environment is a set of resources local to one simulation.
// Rej and Dep are two of its public interfaces (channels), former of which
// outputs rejected customers, while the latter outputs successfully departed
// customers.
type Environment struct {
	acc chan float64  // stream of arrival times of accepted customers
	Rej chan Customer // stream of rejected customers
	srv chan float64  // stream of start servicing times of customers in line
	Dep chan Customer // stream of departed customers
	chl chan float64  // stream of time points when line is available
	chs chan float64  // stream of time points when server is available
	cus Customer      // the current customer
}

// Create one Environment for one simulation.
func NewEnvironment() (e *Environment) {
	e = new(Environment)
	e.acc = make(chan float64)
	e.Rej = make(chan Customer)
	e.srv = make(chan float64)
	e.Dep = make(chan Customer)
	e.chl = make(chan float64)
	e.chs = make(chan float64)
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
func Arrive(e *Environment, rate float64) {
	var now float64
	gen := newExpRNG(rate)
	go func() {
		// accept the 1st customer
		e.acc <- 0.0
		for {
			t := <-e.chl // line will have space at t
			now += gen()
			for now < t {
				e.Rej <- Customer{T0: now}
				now += gen()
			} // now >= t
			e.cus.T0 = now // set arrival time of the current customer
			e.acc <- now   // notice Line the new accepted arrival
		}
	}()
}

// KEY CONCEPT! To be documented.
type clock struct {
	id  int     // each clock has an integer identifier
	now float64 // current time of the clock
}

// A seat is a unit resource of the waiting line.
type seat clock

// A line is seats arranged in queue.
// Note: "queue" is referred here as a FIFO container with fixed capacity. Its
// current implementation resembles ring buffer.
type line struct {
	arr []seat // refers to an underlying array that implements ring buffer
	i   int    // index to the back of the queue (inclusive)
}

func makeline(n int) (q line) {
	q.arr = make([]seat, n)

	// assign identifier for each seat
	for i := 0; i < n; i++ {
		q.arr[i].id = i
	}
	return
}

// Return the succeeding index into the line's underlying array.
// Once the index reaches capacity of the line, reset it to zero.
func (q line) isuc() (i int) {
	i = q.i + 1
	if i == cap(q.arr) {
		i = 0
	}
	return
}

func min(a, b float64) (m float64) {
	m = a
	if m > b {
		m = b
	}
	return
}

// Go manage the FIFO waiting line!
func Line(e *Environment, k int) {
	q := makeline(k)

	go func() {
		// handle the 1st accepted customer
		t0 := <-e.acc
		if q.arr[q.i].now > t0 {
			q.i = q.isuc()
		}
		q.arr[q.i].now = t0
		id := q.arr[q.i].id
		e.cus.T1 = t0
		e.cus.SeatID = id
		e.srv <- t0
		e.chl <- t0

		for {
			t0 = <-e.acc
			if q.arr[q.i].now > t0 {
				q.i = q.isuc()
			}
			q.arr[q.i].now = t0
			id = q.arr[q.i].id

			t1 := <-e.chs
			if t0 > t1 {
				t1 = t0 // time of arrival == time of service
			} else {
				q.arr[q.i].now = t1 // wait until time of service
			}
			e.cus.T1 = t1
			e.cus.SeatID = id
			e.srv <- t1
			e.chl <- min(q.arr[q.i].now, q.arr[q.isuc()].now)
		}
	}()
}

// A server is a unit resource of the server group.
type server struct {
	clock
	gen ExpRNG
}

// A group is servers arranged in min-heap (according to their clock).
type group [](*server) // where heap is built upon

func (h group) swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Given the i-th element within the heap, return the index of the
// minimum of the i-th element, its left child, and its right child.
func (h group) minOfTri(i int) (min int) {
	// return i if h[i] <= h[left] && h[i] <= h[right]
	min = i

	j := 2*i + 1 // index of the left child
	if j > len(h)-1 {
		return
	}
	// j is a valid index
	k := j + 1 // index of the right child
	if k > len(h)-1 {
		if h[i].now > h[j].now {
			min = j
		}
		return
	}
	// k is a valid index
	if h[i].now > h[j].now {
		if h[j].now <= h[k].now {
			min = j
		} else {
			min = k
		}
	} else if h[i].now > h[k].now {
		min = k
	}
	return
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
		h.swap(s, t) // floating down
		s = t
	}
	return
}

// Return the current time of the server on top of the heap.
func (h group) top() (now float64) {
	now = h[0].now
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
func Serve(e *Environment, c int, rate float64) {
	h := makegroup(c, rate)
	go func() {
		// depart the 1st customer
		t1 := <-e.srv
		e.cus.T2, e.cus.SrvrID = h.gen(t1)
		e.Dep <- e.cus
		e.chs <- h.top()
		for {
			t1 = <-e.srv
			e.cus.T2, e.cus.SrvrID = h.gen(t1)
			e.Dep <- e.cus
			e.chs <- h.top()
		}
	}()
}
