package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Customer struct {
	t0	float64
	t1	float64
	t2	float64
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
