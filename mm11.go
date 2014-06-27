package main

import (
	"fmt"
	"time"
)

var acc, rej, srv, dep chan Customer
var chl, chs chan float64

func arrive(gen ExpRng) {
	var now float64
	go func() {
		// accept the 1st customer
		acc <- Customer{t0: now} // now == 0.0 here
		for {
			t := <-chl // line will have space at t
			now += gen()
			for now < t {
				rej <- Customer{t0: now}
				now += gen()
			} // now >= t
			acc <- Customer{t0: now}
		}
	}()
}

func line() {
	//var now float64
	var cus Customer
	go func() {
		// get the 1st accepted customer
		cus = <-acc
		// and serve it
		srv <- cus
		// and "I have space now"
		chl <- 0.0
		for {
			cus = <-acc
			t := <-chs // server will be idle at t
			if cus.t0 > t {
				t = cus.t0
			}
			srv <- Customer{t0: cus.t0, t1: t}
			chl <- t // line will have space at t
		}
	}()
}

func serve(gen ExpRng) {
	var now float64
	var cus Customer
	go func() {
		// depart the 1st customer
		cus = <-srv
		now = cus.t1 + gen()
		dep <- Customer{cus.t0, cus.t1, now}
		chs <- now // server will be idle at now
		for {
			cus = <-srv
			now = cus.t1 + gen()
			dep <- Customer{cus.t0, cus.t1, now}
			chs <- now
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

	arrive(newExpRng(1.0))
	line()
	serve(newExpRng(1.0))
	var cus Customer
	for i := 0; i < 1000; i++ {
		select {
		case cus = <-dep:
			if cus.t0 == cus.t1 {
				fmt.Print("served immediately; ")
			}
			fmt.Println("Departure[", cus.t0, cus.t1, cus.t2, "]")
		case cus = <-rej:
			fmt.Println("Rejection[", cus.t0, "]")
		}
	}

	time.Sleep(2 * time.Second)
}
