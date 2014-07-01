package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Customer struct {
	t0     float64
	t1     float64
	t2     float64
	seat   int
	server int
}

type ExpRng func() float64

func newExpRng(rate float64) ExpRng {
	seed := time.Now().UnixNano()
	fmt.Println("seed: ", seed)
	r := rand.New(rand.NewSource(seed))
	return func() float64 {
		return r.ExpFloat64() / rate
	}
}

type seat struct {
	id  int
	now float64
}

type queueOfSeats struct {
	arr  []seat
	bi int // new comming customer sits at "bi"
}

func (q queueOfSeats) nextIndex() (newback int) {
	newback = (q.bi + 1) % cap(q.arr)
	return
}

func (q *queueOfSeats) sit(c Customer) (sID int, sNow float64) {
	if q.arr[q.bi].now > c.t0 {
		q.bi = q.nextIndex()
	}
	q.arr[q.bi].now = c.t0
	sID = q.arr[q.bi].id
	sNow = q.arr[q.bi].now
	return
}

func (q queueOfSeats) back() (now float64) {
	now = q.arr[q.bi].now
	nownext := q.arr[q.nextIndex()].now
	if now > nownext {
		now = nownext
	}
	return
}

func line_mmxx(n int) {
	var q queueOfSeats
	q.arr = make([]seat, n)
	for i := 0; i < n; i++ {
		q.arr[i].id = i
	}
	var c Customer
	var sID int
	var sNow float64

	go func() {
		// get the 1st accepted customer
		c = <-acc
		sID, sNow = q.sit(c)
		// and serve it
		srv <- Customer{
			t0: sNow,
			t1: 0.0,
			seat: sID,
		}
		// and "I have space now"
		chl <- 0.0
		for {
			c = <-acc
			sID, sNow = q.sit(c)
			t := <-chs
			if sNow > t {
				t = sNow
			} else {
				q.arr[q.bi].now = t
			}
			srv <- Customer{
				t0:   sNow,
				t1:   t,
				seat: sID,
			}
			chl <- q.back()
		}
	}()
}

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
	for i := 0; i < 10000; i++ {
		select {
		case cus = <-dep:
			if cus.t0 == cus.t1 {
				fmt.Print("served immediately; ")
			}
			fmt.Println("Departure[", cus.t0, cus.t1, cus.t2, cus.seat, cus.server, "]")
		case cus = <-rej:
			fmt.Println("Rejection[", cus.t0, "]")
		}
	}

	time.Sleep(2 * time.Second)
}