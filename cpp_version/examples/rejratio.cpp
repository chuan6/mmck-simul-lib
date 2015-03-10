#include <iostream>
#include "mmck.hpp"

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
