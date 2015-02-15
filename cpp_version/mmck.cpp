#include<chrono>
#include<functional>
#include<iostream>
#include<random>
#include<vector>

struct clock {
        int id;
        double epoch; // point in time when an event happens

        clock() : id(0), epoch(0.0) {}
        clock(int id) : id(id), epoch(0.0) {}
};

struct nexter {
        virtual double next() = 0;
        virtual ~nexter() {}
};

struct resource {
        virtual double available_at() const = 0;
        virtual ~resource() {}
};

struct line : resource {
        virtual void wait_or_pass(double, double, double*, int*) = 0;
};

struct service : resource {
        virtual void serve(double, double*, int*) = 0;
};

struct arrival : clock, nexter {};

typedef struct clock seat;

struct server : clock, nexter {};

struct customer {
        double t0;
        double t1;
        double t2;
        int seat_id;
        int server_id;

        customer() : t0(0.0), t1(0.0), t2(0.0), seat_id(0), server_id(0) {}
};

class run {
        arrival* arr;
        line* buf;
        service* srv;

        double chl;
        double chs;
public:
        run(arrival* a, line* l, service* s)
                : arr(a), buf(l), srv(s), chl(0.0), chs(0.0) {};

        customer next() {
                customer cus;
                
                cus.t0 = arr->next();// std::cout<<"hello, " << cus.t0 << std::endl;
                
                if (cus.t0 < chl) return cus;
                // cus.t0 >= chl
                //std::cout << "accepted" << std::endl;

                buf->wait_or_pass(cus.t0, chs, &cus.t1, &cus.seat_id);
                chl = buf->available_at();
                //std::cout << "waited; chl: " << chl << std::endl;

                srv->serve(cus.t1, &cus.t2, &cus.server_id);
                chs = srv->available_at();
                //std::cout << "chs: " << chs << std::endl;

                return cus;
        }
};

typedef std::chrono::high_resolution_clock myclock;

static myclock::time_point beginning = myclock::now();

class exp_arrival : public arrival {
        std::function<double(void)> rng;
public:
        exp_arrival(double r) {
                std::default_random_engine re((myclock::now() - beginning).count());
                std::exponential_distribution<double> exp_dist(r);
                rng = std::bind(exp_dist, re);
        }

        double next() { return (epoch += rng()); }
};

class ring : public line {
        std::vector<seat> buf;
        int back;
public:
        ring(int cap) : buf(cap), back(0) {
                int i = 1;
                for (auto it = buf.begin(); it != buf.end(); it++) {
                        it->id = i++; // assign each seat of the line a sequence number
                }
        }

        double available_at() const { return buf[back].epoch; }

        void wait_or_pass(double t0, double t, double* t1, int* seat_id) {
                /*for (auto it = buf.begin(); it != buf.end(); it++) {
                        std::cout << it->epoch << ", ";
                        }
                        std::cout<<std::endl;*/
                *t1 = buf[back].epoch = (t0 < t)? t : t0;
                *seat_id = buf[back].id; // assign seat_id even when a customer passes the line

                back = (back+1) % buf.size();
        }
};

class exp_server : public server {
        std::function<double(void)> rng;
public:
        exp_server(double r) {
                std::default_random_engine re((myclock::now() - beginning).count());
                std::exponential_distribution<double> exp_dist(r);
                rng = std::bind(exp_dist, re);
        }

        double next() {
                return (epoch += rng());
        }
};

class minheap_service : public service {
        std::vector<server*> heap;

        int min(int i, int j) const {
                return (heap[j]->epoch < heap[i]->epoch)? j : i;
        }
        
        int min_of_tri(int i) const {
                int j = 2*i + 1;
                int k = j + 1;
                int limit = heap.size() - 1;

                if (j > limit)
                        return i;
                if (k > limit)
                        return min(i, j);
                if (min(i, j) == i)
                        return min(i, k);
                return min(j, k);
        }
public:
        minheap_service(std::vector<server*> init) : heap(init) {}

        double available_at() const { return heap[0]->epoch; }
        
        void serve(double t1, double* t2, int* server_id) {
                heap[0]->epoch = t1;
                *t2 = heap[0]->next();
                *server_id = heap[0]->id;

                server* tmp;
                int s = 0;
                for (int t = min_of_tri(s); s != t; t = min_of_tri(s)) {
                        tmp = heap[s];
                        heap[s] = heap[t];
                        heap[t] = tmp;
                        s = t;
                }
        }
};

int main() {
        exp_arrival* arr = new exp_arrival(2.0);
        ring* line = new ring(5);
        std::vector<exp_server*> expservers = {new exp_server(1.0), new exp_server(1.0)};
        int i = 0;
        for (auto it=expservers.begin(); it != expservers.end(); it++, i++) {
                (*it)->id = i;
        }
        std::vector<server*> servers(expservers.begin(), expservers.end());
        minheap_service* srv = new minheap_service(servers);
        
        run r(arr, line, srv);
        customer c;
        int rej = 0, dep = 0;
        double staytime = 0.0;
        for (int i = 0; i < 100000000; i++) {
                c = r.next();
                /*std::cout << "t0: " << c.t0
                          << "\tt1: " << c.t1
                          << "\tt2: " << c.t2
                          << "\tseat: " << c.seat_id
                          << "\tserver: " << c.server_id << std::endl;*/

                if (c.t2 == 0.0) {
                        rej++;
                } else {
                        dep++;
                        staytime += c.t2 - c.t0;
                }
        }
        std::cout << "rej: " << rej
                  << " dep: " << dep
                  << " mean stay time: " << staytime/(double)dep << std::endl;

        delete arr;
        delete line;
        delete srv;
}
