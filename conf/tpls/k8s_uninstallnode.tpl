- hosts:
  - k8s-worker
  - k8s-worker-new
  vars:
    DEL_MASTER: "yes"
    DEL_NODE: "yes"
    DEL_ETCD: "yes"
    DEL_LB: "yes"
    DEL_CHRONY: "yes"
    DEL_ENV: "yes"
  gather_facts: no
  roles:
  - k8s-del-node
  - clean