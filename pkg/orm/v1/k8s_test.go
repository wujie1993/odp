package v1_test

import (
	"encoding/json"
	"testing"

	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/orm/core"
	v1 "github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/setting"
)

func init() {
	setting.EtcdSetting = &setting.Etcd{
		Endpoints: []string{core.DefaultEtcdEndpoint},
	}
	db.InitKV()
}

func TestK8sYamlEncoding(t *testing.T) {
	host := v1.NewV1K8sYaml()

	bytes, err := json.MarshalIndent(host, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("encode: %s", string(bytes))

	if err := json.Unmarshal(bytes, host); err != nil {
		t.Fatal(err)
	}
	t.Logf("decode: %+v", *host)
}

func TestK8sYamlCRUD(t *testing.T) {
	k8syaml := v1.NewV1K8sYaml()
	k8syaml.Metadata.Namespace = "default"
	k8syaml.Metadata.Name = "host1"
	k8syaml.Spec.K8SMaster.Hosts = []v1.HostRef{}
	k8syaml.Spec.K8SMaster.Hosts = append(k8syaml.Spec.K8SMaster.Hosts, v1.HostRef{
		ValueFrom: v1.V1ValueFrom{
			HostRef: "hsot-234",
		},
	})
	k8syaml.Spec.K8SWorker.Hosts = append(k8syaml.Spec.K8SMaster.Hosts, v1.HostRef{
		ValueFrom: v1.V1ValueFrom{
			HostRef: "hsot-235",
		},
	})
	// 也可使用 helper := orm.GetHelper()
	registry := v1.NewV1K8sYamlRegistry()

	// 创建V1Host，也可使用 helper.V1.Host.Create(host) 或 helper.V1.Host.CreateHost(host)
	if h, err := registry.Create(k8syaml); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("create k8syaml succeed: %s", string(data))
		}
	}

	// 获取V1Host，也可使用 helper.V1.Host.Get(host) 或 helper.V1.Host.GetHost(host)
	if h, err := registry.Get(k8syaml.Metadata.Namespace, k8syaml.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("get k8syaml succeed: %s", string(data))
		}
	}

	// k8syaml.Spec.K8SMaster.Hosts.hostRef ="172.25.21.25"
	// // 更新V1Host，也可使用 helper.V1.Host.Update(host) 或 helper.V1.Host.UpdateHost(host)
	// if h, err := registry.Update(k8syaml); err != nil {
	// 	t.Error(err)
	// } else {
	// 	if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
	// 		t.Error(err)
	// 	} else {
	// 		t.Logf("update k8syaml succeed: %s", string(data))
	// 	}
	// }

	// if h, err := registry.Get(k8syaml.Metadata.Namespace, k8syaml.Metadata.Name); err != nil {
	// 	t.Error(err)
	// } else if h.(*v1.V1K8sClusterYaml).Spec.K8SMaster.Hosts.hostRef !="172.25.21.25"{
	// 	t.Error("update failed: result not equal to what you updated")
	// }

	// 列举V1Host，也可使用 helper.V1.Host.List(host.Metadata.Namespace) 或 helper.V1.Host.ListHost(host.Metadata.Namespace)
	if list, err := registry.List(k8syaml.Metadata.Namespace); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(list, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("list k8syaml succeed: %s", string(data))
		}
	}

	// 删除V1Host，也可使用 helper.V1.Host.Delete(host.Metadata.Namespace, host.Metadata.Name) 或 helper.V1.Host.DeleteHost(host.Metadata.Namespace, host.Metadata.Name)
	if h, err := registry.Delete(k8syaml.Metadata.Namespace, k8syaml.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("delete k8syaml succeed: %s", string(data))
		}
	}

	if h, err := registry.Get(k8syaml.Metadata.Namespace, k8syaml.Metadata.Name); err != nil {
		t.Error(err)
	} else if h != nil {
		t.Error("delete failed: object still exist")
	}
}
