- hosts:
  - k8s-master
  - k8s-worker
  gather_facts: yes
  roles:
  - { role: set_hostname }

- hosts: localhost
  roles:
  - deploy

- hosts:
  - k8s-master
  - k8s-worker
  roles:
  - { role: prepare, tags: ['prepare'] }

- hosts: etcd
  roles:
  - etcd
  tags: etcd

- hosts:
  - k8s-master
  - k8s-worker
  - localhost
  roles:
  - { role: docker, when: "CONTAINER_RUNTIME == 'docker'", tags: [docker]}

- hosts: k8s-master
  roles:
  - kube-master
  - kube-node
- hosts:
  - k8s-master
  roles:
  - { role: kube-master-label }


- hosts: k8s-worker
  roles:
  - { role: kube-node, when: "inventory_hostname not in groups['k8s-master']", tags: ['k8s-worker']}

- hosts: k8s-master
  roles:
    - namespaces

- hosts:
  - k8s-master
  - k8s-worker
  roles:
  - { role: calico, when: "CLUSTER_NETWORK == 'calico'", tags: ['calico']}
  - { role: flannel, when: "CLUSTER_NETWORK == 'flannel'", tags: ['flannel']}


- hosts:
  - k8s-master
  roles:
  - { role: cluster-addon, tags: ['addon'] }



- hosts: etcd
  gather_facts: no
  roles:
  - { role: etcd_backup }

  # 11.harbor.yml
- hosts: harbor
  gather_facts: no
  roles:
  - { role: harbor }
  
- hosts:
  - k8s-master
  - k8s-worker
  roles:
  - { role: set_harbor_domain }