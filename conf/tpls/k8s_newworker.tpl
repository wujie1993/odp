- hosts: k8s-worker-new
  gather_facts: yes
  roles:
  - set_hostname
  - prepare
  - docker
  - kube-node
  - { role: calico, when: "CLUSTER_NETWORK == 'calico'", tags: ['calico']}
  - { role: flannel, when: "CLUSTER_NETWORK == 'flannel'", tags: ['flannel']}
  - set_harbor_domain