- hosts:
  - k8s-master
  - k8s-worker
  vars:
    DEL_MASTER: "yes"
    DEL_NODE: "yes"
    DEL_ETCD: "yes"
    DEL_LB: "yes"
    DEL_CHRONY: "yes"
    DEL_ENV: "yes"
  gather_facts: no
  roles:
  - clean