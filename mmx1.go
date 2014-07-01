package main

import (
	//"fmt"
)

type server struct {
	id  int
	now float64
	gen ExpRng
}

type heapOfServers [](*server) // where heap is built upon

func (h heapOfServers) swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h heapOfServers) minOfTri(i int) (min int) {
	// returns i if h[i] <= h[left] && h[i] <= h[right]
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

func (h heapOfServers) gen(now float64) (depTime float64, sid int) {
	sid = h[0].id
	depTime = now + h[0].gen()
	h[0].now = depTime

	s := 0
	for t := h.minOfTri(s); s != t; t = h.minOfTri(s) {
		//fmt.Print("s:", s, "; t:", t, " ")
		h.swap(s, t) // floating down
		s = t
	} // heap property is kept
	return
}

func (h heapOfServers) top() (now float64) {
	now = h[0].now
	return
}

func gensToServers(gens []ExpRng) (h heapOfServers) {
	n := len(gens)
	h = make([]*server, n)
	p := make([]server, n) // pointer to the underlying array that stores servers

	for i := 0; i < n; i++ {
		p[i].id = i
		p[i].gen = gens[i]
		h[i] = &p[i]
	}

	return
}

func serve_mmx1(gens []ExpRng) {
	h := gensToServers(gens)
	var cus Customer
	go func() {
		// depart the 1st customer
		cus = <-srv
		deptime, sid := h.gen(cus.t1)
		dep <- Customer{t0: cus.t0, t1: cus.t1, t2: deptime, seat: cus.seat, server: sid}
		chs <- h.top()
		for {
			cus = <-srv
			deptime, sid = h.gen(cus.t1)
			dep <- Customer{t0: cus.t0, t1: cus.t1, t2: deptime, seat: cus.seat, server: sid}
			chs <- h.top()
		}
	}()
}

/*
func main() {
	acc = make(chan Customer)
	rej = make(chan Customer)
	srv = make(chan Customer)
	dep = make(chan Customer)
	chl = make(chan float64)
	chs = make(chan float64)

	ns := 5
	gensForServer := make([]ExpRng, ns)
	for i := 0; i < ns; i++ {
		gensForServer[i] = newExpRng(1.0)
	}

	arrive_mm11(newExpRng(5.0))
	line_mm11()
	serve_mmx1(gensForServer)
	var cus Customer
	for i := 0; i < 1000; i++ {
		select {
		case cus = <-dep:
			if cus.t0 == cus.t1 {
				fmt.Print("served immediately; ")
			}
			fmt.Println("Departure[", cus.t0, cus.t1, cus.t2, cus.server, "]")
		case cus = <-rej:
			fmt.Println("Rejection[", cus.t0, "]")
		}
	}

	time.Sleep(2 * time.Second)
}
*/
