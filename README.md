# simulation-queueing-mmx
A library for intuitive M/M/c/K queueing system simulation, written in Go.

To use it, 1) arrival rate; 2) queue capacity; 3) service rate per server; and 4) number of servers, need to be provided.

Performance of the queuing system under your configuration is evaluated by "asking" the customers leaving the system (rejected or serviced), about their arrival time, service time, departure time, even the seat No. and server No. the customer used.

Usage:
```go
envptr := NewEnvironment()
Arrive(envptr, _arrival_rate_)
Line  (envptr, _line_capacity_)
Serve (envptr, _number_of_servers_, _service_rate_per_server_)

// statistical analysis on output from the rejection channel, and
// departure channel
var c Customer
for i := 0; i < _nloops_; i++ {
	select {
	case c = <-envptr.Dep:
		// ask the departed customer c
	caes c = <-envptr.Rej:
		// ask the rejected customer c
	}
}
```
