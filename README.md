# mmck-queueing-simulation-library
A library for intuitive M/M/c/K queueing system simulation, implemented both in go and in C++.

(Personal note: I had tried event list based approach, and state machine based approach. None offered the program clarify, modularity that I wanted to achieve.)

A client program can view the library as a "random queueing event generator", to which each call returns a departure or rejection structure. Within the structure, 1) arrival time, 2) service time, 3) departure time, 4) waiting position in queue, and 5) server id are recorded for a single customer.

Besides using default queueing sytem setting, where we have Poisson arrival, FIFO queue, and exponential service time, a client program can provide customized modules for arrival, line (i.e. queue), and service, to simulate systems for which it is difficult to obtain closed-form analytical results, as long as the corresponding interfaces for each module are implemented. For example, a client program can provide service module where each server has different service rate, or even distribution, or line module which incorporates certain priorities other than arrival time. 

Performance-wise, for default queueing system setting, time is mostly spent in calls to random number generators. A million departures+rejections are expected to take much less than one second for go version, and much less than half a second for C++ version. (Tested on my Thinkpad T420s i5 machine.)

To use it, 1) arrival rate; 2) queue capacity; 3) service rate per server; and 4) number of servers, need to be provided.

Performance of the queuing system under your configuration is evaluated by "asking" the customers leaving the system (rejected or serviced), about their arrival time, service time, departure time, and server No. the customer used.

Usage:

go version:

```go
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
```

C++ version:

```c++
int count_rejs(simulation& simul, int narrs) {
    int n = 0;
    customer cus;
    for (int i = 0; i < narrs; i++) {
        cus = simul.next();
        if (is_rejected(cus))
            n++;
        }
    return n;
}

int main() {
    exp_arrival arr(2.0);
    ring buf(5);
    std::vector<exp_server> exp_srvrs {exp_server(1, 1.0), exp_server(2, 1.0)};
    std::vector<server*> srvrs(exp_srvrs.size());
    auto p = srvrs.begin();
    for (auto q = exp_srvrs.begin(); p != srvrs.end(); p++, q++) {
        *p = exp_srvrs.data() + (q - exp_srvrs.begin());
    }
    minheap_service srv(srvrs);
    simulation simul(arr, buf, srv);
    int n = 100000000;
    int nrej = count_rejs(simul, n);
    std::cout << "ratio: " << (double) nrej / n << std::endl;
}
```
Design:
![Alt text](images_design_illustration/scan1.jpg?raw=true "Page 1.")
![Alt text](images_design_illustration/scan2.jpg?raw=true "Page 2.")
