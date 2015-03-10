#ifndef chuan6_mmck_h
#define chuan6_mmck_h

#include <chrono>
#include <functional>
#include <iostream>
#include <random>
#include <vector>

// the complete record of experience of a customer
struct customer {
    double t0;        // time of arrival
    double t1;        // time of service
    double t2;        // time of departure
    int    seat_id;   // position where the customer waited in line
    int    server_id; // server that served the customer
};
// decide if a generated customer is rejected
bool is_rejected(const customer&);
// pretty print a customer structure
std::ostream& operator<<(std::ostream&, const customer&);

// the arrival module, each seat in the line module, and each server in the
// service module keep a clock, as their internal state 
struct clock {
    int    id;    // for easy identification
    double epoch; // the time when an event happens

    clock() : id(0), epoch(0.0) {}
    clock(int id) : id(id), epoch(0.0) {}
};

// the line module represents a resource whose earliest available time is of
// direct concern to the arrival module (e.g. determine if an arrival is to be
// rejected); similarly, the service module represents a resource whose
// earliest available time is of direct concern to the line module (e.g.
// determine if an arrival is to simply pass through or wait, and for how long)
struct resource {
    virtual double earliest_available() const = 0;
    virtual ~resource() {}
};

// the arrival module, and each server in the service module self-forward their
// clocks to update states
struct progress {
    virtual double forward() = 0;
    virtual ~progress() {}
};

// the arrival module (default: exponential arriving interval)
struct arrival : clock, progress {};

// the line module (default: fixed length, single FIFO queue)
struct line : resource {
    // input: arrival time, and earliest available time of service
    // output: service time, and position where the customer waited in line
    // note: service time is the scheduled time to begin service
    virtual void wait_or_pass(double, double, double&, int&) = 0;
};

// the service module (default: minheap, i.e. assign earliest available server
// for the next customer from the line)
struct service : resource {
    // input: service time
    // output: departure time, and id of the server that served the customer
    virtual void serve(double, double&, int&) = 0;
};

typedef struct clock seat; // a seat in waiting line is simply a clock

struct server : clock, progress {
    server(int id) : clock(id) {}
};

// The main algorithm
class simulation {
    arrival& arr; // the arrival module,
    line&    buf; // the line module,
    service& srv; // and the service module
    // that are configured for this simulation
public:
    simulation(arrival& a, line& l, service& s) : arr(a), buf(l), srv(s) {}

    // each call to next() generates a complete record of experience of a
    // customer according to the configured arrival, line, and service modules
    customer next() {
        customer cus{}; // zero-initialized

        cus.t0 = arr.forward();
        double chl = buf.earliest_available();
        if (cus.t0 < chl)
            return cus; // rejected
        // accepted

        double chs = srv.earliest_available();
        buf.wait_or_pass(cus.t0, chs, cus.t1, cus.seat_id);
        // waited

        srv.serve(cus.t1, cus.t2, cus.server_id);
        // served
        
        return cus;
    }
};

// the default arrival module, that generates arrivals with exponential
// arriving interval
class exp_arrival : public arrival {
    std::function<double(void)> gen;
public:
    exp_arrival(double); // input: arrival rate
    double forward() { return (epoch += gen()); }
};

// the default line module, implemented as fixed length ring buffer
class ring : public line {
    std::vector<seat> buf;
    int len;  // length of buf; use len instead of buf.size()
    int back; // buf[back] denotes the seat for next arrival
public:
    ring(size_t n) : buf(n), len(n), back(0) {}

    void wait_or_pass(double t0, double t, double& t1, int& sid) {
        t1 = buf[back].epoch = (t0 < t)?
            t : // wait til t
            t0; // or pass through
        sid = buf[back].id; // assign seat id even for pass through cases
        back = (back+1) % len;
    }

    // the property that the earliest available time of this line module
    // equals buf[back].epoch is not imposed by this line implementation;
    // instead, it is imposed by the minheap service module.
    // note: a more expensive wait_or_pass function can be implemented
    // to guarantee this property regardless of service module implementation,
    // but here, for the default setting, this algorithm works fine.
    double earliest_available() const {
        return buf[back].epoch;
    }
};

// the default server implementation that generates exponential durations
// of service
class exp_server : public server {
    std::function<double(void)> gen;
public:
    exp_server(int, double);

    double forward() { return (epoch += gen()); }
};

class minheap_service : public service {
    std::vector<server*> heap;
    int limit; // greatest index in heap

    int min(int i, int j) const; // used by heapify
    int min_of_tri(int i) const; // used by heapify
    void heapify(int i);
public:
    minheap_service(std::vector<server*> init) : heap(init), limit(init.size()-1) {
        for (int i = (limit-1)/2; i >= 0; i--) {
            heapify(i);
        }
    }

    void serve(double t1, double& t2, int& sid) {
        heap[0]->epoch = t1;
        t2 = heap[0]->forward();
        sid = heap[0]->id;
        heapify(0);
    } 

    // the property that the earliest available time of this service module
    // equals heap[0]->epoch is imposed by the minheap implementation
    double earliest_available() const { return heap[0]->epoch; }
};

#endif
