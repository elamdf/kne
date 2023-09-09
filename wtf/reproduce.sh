#!/bin/bash
set -e
./deploy_topo_app.sh


print_and_run() {
    echo "Running command: $@"
    eval "$@"

}

print_and_run kubectl -n wtf exec -it egress -- ip addr | grep 192
print_and_run kubectl -n wtf exec -it srl0 -- ip netns exec srbase-DEFAULT ping 192.168.0.5 -c1 -I192.168.0.0
print_and_run kubectl -n wtf exec -it srl1 -- ip netns exec srbase-DEFAULT ping 192.168.0.5 -c1 -I192.168.0.2
print_and_run kubectl -n wtf exec -it srl2 -- ip netns exec srbase-DEFAULT ping 192.168.0.5 -c1 -I192.168.0.4
print_and_run kne topology push out/wtf_topo.pbtxt srl2 out/loop_create_srl2_src_srl0.cfg
print_and_run kubectl -n wtf exec -it srl0 -- ip netns exec srbase-DEFAULT curl --interface 192.168.0.0 http://172.18.0.50:80/productpage -H 'x-request-id: test-request-srl0-loop'
print_and_run kne topology push out/wtf_topo.pbtxt srl2 out/loop_undo_srl2_src_srl0.cfg
print_and_run kubectl -n wtf exec -it srl0 -- ip netns exec srbase-DEFAULT curl --interface 192.168.0.0 http://172.18.0.50:80/productpage -H 'x-request-id: test-request-srl0-loop-undone'
print_and_run kubectl -n wtf exec -it srl0 -- ip netns exec srbase-DEFAULT curl --interface 192.168.0.0 http://172.18.0.50:80/productpage -H 'x-request-id: WTFTRACE-test-request-srl0-loop'