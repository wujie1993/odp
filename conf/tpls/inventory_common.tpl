all:
  vars:
    CLUSTER_CIDR: 10.244.0.0/16
    CLUSTER_DNS_DOMAIN: cluster.local.
    CLUSTER_NETWORK: calico
    CONTAINER_RUNTIME: docker
    HARBOR_DOMAIN: harbor.pcitech.com
    NODE_PORT_RANGE: 20000-40000
    PROXY_MODE: ipvs
    SERVICE_CIDR: 10.68.0.0/16
    base_dir: {{ .BaseDir }}
    bin_dir: /opt/kube/bin
    ca_dir: /etc/kubernetes/ssl
    manifest_dir: /opt/prophet/manifests
    prophet_version: v1.3