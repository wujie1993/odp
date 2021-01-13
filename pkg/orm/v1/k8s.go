package v1

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
)

type K8sConfig struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            K8sYaml
}

type K8sYaml struct {
	Action string
	Chrony string
	Etcd   struct {
		Hosts []K8sHostRef
	}
	ExLb   string
	Harbor struct {
		Hosts []K8sHostRef
	}
	K8SMaster struct {
		Hosts []K8sHostRef
	} `json:"K8s-master" yaml:"K8s-master"`
	K8SWorker struct {
		Hosts []K8sHostRef
	} `json:"K8s-worker" yaml:"K8s-worker"`
	K8SWorkerNew struct {
		Hosts []K8sHostRef
	} `json:"K8s-worker-new" yaml:"K8s-worker-new"`
	GPU struct {
		Hosts []K8sHostRef
	}
}

type K8sHostRef struct {
	ValueFrom ValueFrom
	Label     map[string]string
}

func (obj K8sConfig) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

func (obj *K8sConfig) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

func (obj K8sConfig) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// +namespaced=true
type K8sConfigRegistry struct {
	registry.Registry
}

func (r K8sConfigRegistry) GetFirstMasterHost(name string) (*Host, error) {
	// 获取k8s集群节点
	k8sObj, err := r.Get(context.TODO(), core.DefaultNamespace, name)
	if err != nil {
		return nil, err
	} else if k8sObj == nil {
		return nil, e.Errorf("k8s cluster '%s' not found", name)
	}
	k8s := k8sObj.(*K8sConfig)

	if len(k8s.Spec.K8SMaster.Hosts) < 1 {
		return nil, e.Errorf("k8s cluster '%s' does not have any master hosts exist", name)
	}
	// 获取k8s集群第一个master节点
	hostRef := k8s.Spec.K8SMaster.Hosts[0].ValueFrom.HostRef
	hostRegistry := NewHostRegistry()
	hostObj, err := hostRegistry.Get(context.TODO(), "", hostRef)
	if err != nil {
		return nil, err
	} else if hostObj == nil {
		return nil, e.Errorf("host %s not found", hostRef)
	}
	return hostObj.(*Host), nil
}

func NewK8sConfig() *K8sConfig {
	k8sConfig := new(K8sConfig)
	k8sConfig.Init(ApiVersion, core.KindK8sConfig)
	return k8sConfig
}

func NewK8sConfigRegistry() K8sConfigRegistry {
	r := K8sConfigRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindK8sConfig), true),
	}
	r.SetDefaultFinalizers([]string{
		core.FinalizerCleanRefEvent,
	})
	return r
}
