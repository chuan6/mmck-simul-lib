package mmck

/*
Sample usage:
    var rejected, departed <-chan mmck.Customer
    rejected, departed = mmck.Run(
        mmck.NewExpArrival(10.0),
        mmck.NewFifoLine(7),
        mmck.MakeMinheapExpService(2, 1.0),
    )
    var cus mmck.Customer
    for i := 0; i < _n_arrivals_; i++ {
        select {
        case cus = <-rejected:
            // do statistics for rejected customers
        case cus = <-departed:
            // do statistics for departed customers
        }
    }
*/

type Customer struct {
	T0     float64 // time of arrival of the customer
	T1     float64 // time of service of the customer
	T2     float64 // time of departure of the customer
	SeatID int     // identifier of the seat that sits the waiting customer
	SrvrID int     // identifier of the server that serves the customer
}

type Arrival interface {
	Arrive() (at float64)
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

type Line interface {
	Waiter
	Nexter
}

type Service interface {
	Server
	Nexter
}

func Run(a Arrival, l Line, s Service) (rejs, deps <-chan Customer) {
	var chl, chs float64
	rej := make(chan Customer, 8)
	dep := make(chan Customer, 32)

	go func() {
		dep <- Customer{}

		var t0, t1, t2 float64
		var seat, server int
		for {
			t0 += a.Arrive()
			for t0 < chl {
				rej <- Customer{T0: t0}
				t0 += a.Arrive()
			} // t0 >= chl
			// accepted, or rejected

			t1, seat = l.WaitOrPass(t0, chs)
			chl = l.Next()
			// waited

			t2, server = s.Serve(t1)
			// served

			dep <- Customer{
				T0: t0,
				T1: t1,
				T2: t2,
				SeatID: seat,
				SrvrID: server,
			}
			chs = s.Next()
		}
	}()

	rejs = rej
	deps = dep
	return
}

