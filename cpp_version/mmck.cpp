#include <chrono>
#include <functional>
#include <iostream>
#include <random>
#include <vector>

#define NDEBUG

#include <cassert>

struct customer {
        double t0;     // time of arrival
        double t1;     // time of service
        double t2;     // time of departure
        int seat_id;   // where the customer waited in line
        int server_id; // where the customer got service
};

bool is_rejected(const customer& cus) {
        return cus.t1 < cus.t0;
}

std::ostream& operator<<(std::ostream& out, const customer& cus) {
        out << "t0: " << cus.t0 << "\t"
            << "t1: " << cus.t1 << "\t"
            << "t2: " << cus.t2 << "\t"
            << "seat_id: " << cus.seat_id << "\t"
            << "server_id: " << cus.server_id << "\t";
        return out;
}

struct clock {
        int id;
        double epoch;

        clock() : id(0), epoch(0.0) {}
        clock(int id) : id(id), epoch(0.0) {}
};

struct resource {
        virtual double earliest_available() const = 0;
        virtual ~resource() {}
};

struct progress {
        virtual double forward() = 0;
        virtual ~progress() {}
};

struct arrival : clock, progress {};

struct line : resource {
        virtual void wait_or_pass(double, double, double&, int&) = 0;
};

struct service : resource {
        virtual void serve(double, double&, int&) = 0;
};

typedef struct clock seat;

struct server : clock, progress {
        server(int id) : clock(id) {}
};

// The main algorithm
class simulation {
        arrival& arr;
        line& buf;
        service& srv;
public:
        simulation(arrival& a, line& l, service& s) : arr(a), buf(l), srv(s) {}

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

typedef std::chrono::high_resolution_clock myclock;

static myclock::time_point beginning = myclock::now();

static std::function<double(void)> get_expgen(double rate) {
        std::default_random_engine re((myclock::now() - beginning).count());
        std::exponential_distribution<double> exp(rate);
        return std::bind(exp, re);
}

class exp_arrival : public arrival {
        std::function<double(void)> gen;
public:
        exp_arrival(double rate) : gen(get_expgen(rate)) {}

        double forward() { return (epoch += gen()); }
};

class ring : public line {
        std::vector<seat> buf;
        int len;
        int back;
public:
        ring(size_t n) : buf(n), len(n), back(0) {}

        void wait_or_pass(double t0, double t, double& t1, int& sid) {
                t1 = buf[back].epoch = (t0 < t)? t : t0;
                sid = buf[back].id;
                back = (back+1) % len;
        }
        
        double earliest_available() const { return buf[back].epoch; }
};

class exp_server : public server {
        std::function<double(void)> gen;
public:
        exp_server(int id, double rate) : server(id), gen(get_expgen(rate)) {}

        double forward() { return (epoch += gen()); }
};

class minheap_service : public service {
        std::vector<server*> heap;
        int limit; // greatest index in heap

        int min(int i, int j) const {
                return (heap[i]->epoch > heap[j]->epoch)? j : i; 
        }
        int min_of_tri(int i) const {
                int j = 2*i + 1;
                if (j > limit || j < 0)
                        return i;
                
                int k = j + 1;
                if (k > limit || k < 0)
                        return min(i, j);

                // i, j, k are all valid index
                if (i == min(i, j))
                        return min(i, k);
                return min(j, k);
        }
        void heapify(int i) {
                server* tmp;
                for (int x = min_of_tri(i); x != i; x = min_of_tri(i)) {
                        tmp = heap[i];
                        heap[i] = heap[x];
                        heap[x] = tmp;
                        i = x;
                }
        }
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

        double earliest_available() const { return heap[0]->epoch; }
};

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
