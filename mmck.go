package mmck

// the complete record of experience of a customer
type Customer struct {
	T0     float64 // time of arrival of the customer
	T1     float64 // time of service of the customer
	T2     float64 // time of departure of the customer
	SeatID int     // identifier of the seat that sits the waiting customer
	SrvrID int     // identifier of the server that serves the customer
}

// a customized arrival module needs to implement Arrival interface
type Arrival interface {
	ArriveIn() (at float64)
}

type Waiter interface {
	WaitOrPass(arriveAt float64, serverAvailableAt float64) (serveAt float64, seatID int)
}

type Server interface {
	Serve(startAt float64) (departAt float64, serverID int)
}

type Nexter interface {
	Next() (nextAvailableAt float64)
}

// a customized queue module needs to implement Line interface
type Line interface {
	Waiter
	Nexter
}

// a customized service module needs to implement Service interface
type Service interface {
	Server
	Nexter
}

func Run(a Arrival, l Line, s Service) (rejs, deps <-chan Customer) {
	rej := make(chan Customer, 8)
	dep := make(chan Customer, 32)
	// Note: adjust buffer size of both channels to tune performance;
	// default value works best in my cases.

	go func() {
		var t0, t1, t2 float64
		var seat, server int
		var chl, chs float64 // time of next available for line, and service
		for {
			t0 += a.ArriveIn()
			for t0 < chl {
				rej <- Customer{T0: t0}
				t0 += a.ArriveIn()
			} // zero or more rejected, one accepted

			t1, seat = l.WaitOrPass(t0, chs)
			chl = l.Next()
			// waited from t0 to t1 at seat

			t2, server = s.Serve(t1)
			chs = s.Next()
			// served from t1 to t2 by server

			dep <- Customer{
				T0:     t0,
				T1:     t1,
				T2:     t2,
				SeatID: seat,
				SrvrID: server,
			} // departed
		}
	}()

	rejs = rej
	deps = dep
	return
}
