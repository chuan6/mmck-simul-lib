# simulation-queueing-mmx
A library for intuitive M/M/c/K queueing system simulation, written in Go. The implementation is designed so that user can easily give different service rate for many servers, which is a difficult problem to get analytical solution.

To use it, 1) arrival rate; 2) queue capacity; 3) service rate per server; and 4) number of servers, need to be provided.

Performance of the queuing system under your configuration is evaluated by "asking" the customers leaving the system (rejected or serviced), about their arrival time, service time, departure time, even the seat No. and server No. the customer used.

Usage:
```go
simtask := mmx.NewEnvironment()
mmx.Arrive(simtask, _arrival_rate_)
mmx.Line  (simtask, _line_capacity_)
mmx.Serve (simtask, _number_of_servers_, _service_rate_per_server_)

// do statistical analysis on output from the rejection channel, and
// departure channel
var c mmx.Customer
for i := 0; i < _narrivals_; i++ {
	select {
	case c = <-simtask.Dep:
		// ask the departed customer c
	caes c = <-simtask.Rej:
		// ask the rejected customer c
	}
}
```
