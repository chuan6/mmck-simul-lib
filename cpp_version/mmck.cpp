#include <iostream>

#include "mmck.hpp"

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

typedef std::chrono::high_resolution_clock myclock;
static myclock::time_point beginning = myclock::now();
static std::function<double(void)> get_expgen(double rate) {
    return std::bind(std::exponential_distribution<double>{rate},
                     std::default_random_engine{
                         (unsigned)(myclock::now() - beginning).count()});
}

exp_arrival::exp_arrival(double rate)
    : gen(get_expgen(rate)) {}

exp_server::exp_server(int id, double rate)
    : server(id), gen(get_expgen(rate)) {}

int minheap_service::min(int i, int j) const {
    return (heap[i]->epoch > heap[j]->epoch)? j : i;
}

int minheap_service::min_of_tri(int i) const {
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

void minheap_service::heapify(int i) {
    server* tmp;
    for (int x = min_of_tri(i); x != i; x = min_of_tri(i)) {
        tmp = heap[i];
        heap[i] = heap[x];
        heap[x] = tmp;
        i = x;
    }
}

