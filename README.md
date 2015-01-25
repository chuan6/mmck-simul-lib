# simulation-queueing-mmx
A library for intuitive M/M/c/K queueing system simulation, written in Go.

(Personal note: I had tried implementing the program in both C and C++, using event list based approach, and state machine based approach. None offered the program clarify, modularity  that I wanted to achieve.

In this program, ARRIVE, LINE, and SERVE are separated into three "independent" goroutines, synchronizing only through four channels. The channels carry time points to keep each goroutine moving forward in terms of time of a simulation environment. One difficulty I've encountered is that this approach makes the synchronization explicit, while using event list or state machine hides it.

The resulting program also implements API as output channels, so that a user program can have independence and "real time" simulation updates, both of which are difficult to achieve through other approaches.)

The implementation is designed so that user can easily give different service rate for many servers, which is a difficult problem to get analytical solution.

To use it, 1) arrival rate; 2) queue capacity; 3) service rate per server; and 4) number of servers, need to be provided.

Performance of the queuing system under your configuration is evaluated by "asking" the customers leaving the system (rejected or serviced), about their arrival time, service time, departure time, and server No. the customer used.

Usage:
```go
// do statistical analysis on output from the rejection channel, and
// departure channel
simtask := mmx.NewEnvironment()

simtask.Arrive(_arrival_rate_)
simtask.Line  (_line_capacity_)
simtask.Serve (_number_of_servers_, _service_rate_per_server_)

rejected, departed := simtask.Output()
var c mmx.Customer
for i := 0; i < _ncustomers_; i++ {
        select {
        case c = <-departed:
                // ask the departed customer c
        case c = <-rejected:
                // ask the rejected customer c
        }
}
```
Design:
![Alt text](images_design_illustration/scan1.jpg?raw=true "Page 1.")
![Alt text](images_design_illustration/scan2.jpg?raw=true "Page 2.")
