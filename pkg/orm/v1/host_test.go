package v1_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/wujie1993/waves/pkg/db"
	v1 "github.com/wujie1993/waves/pkg/orm/v1"

	// "github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/setting"
)

func init() {
	setting.EtcdSetting = &setting.Etcd{
		Endpoints: []string{core.DefaultEtcdEndpoint},
	}
	db.InitKV()
}

func TestHostEncoding(t *testing.T) {
	host := v1.NewV1Host()

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

func TestHostCRUD(t *testing.T) {
	host := v1.NewV1Host()
	host.Metadata.Namespace = "default"
	host.Metadata.Name = "host1"
	host.Spec.SSH.Host = "192.168.1.2"
	host.Spec.SSH.Port = 22
	host.Spec.SSH.User = "root"
	host.Spec.SSH.Password = "123456"

	// 也可使用 helper := orm.GetHelper()
	registry := v1.NewV1HostRegistry()

	// 创建V1Host，也可使用 helper.V1.Host.Create(host) 或 helper.V1.Host.CreateHost(host)
	if h, err := registry.Create(host); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("create host succeed: %s", string(data))
		}
	}

	// 获取V1Host，也可使用 helper.V1.Host.Get(host) 或 helper.V1.Host.GetHost(host)
	if h, err := registry.Get(host.Metadata.Namespace, host.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("get host succeed: %s", string(data))
		}
	}

	host.Spec.SSH.Host = "172.17.32.1"
	// 更新V1Host，也可使用 helper.V1.Host.Update(host) 或 helper.V1.Host.UpdateHost(host)
	if h, err := registry.Update(host, false); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("update host succeed: %s", string(data))
		}
	}

	if h, err := registry.Get(host.Metadata.Namespace, host.Metadata.Name); err != nil {
		t.Error(err)
	} else if h.(*v1.V1Host).Spec.SSH.Host != "172.17.32.1" {
		t.Error("update failed: result not equal to what you updated")
	}

	// 列举V1Host，也可使用 helper.V1.Host.List(host.Metadata.Namespace) 或 helper.V1.Host.ListHost(host.Metadata.Namespace)
	if list, err := registry.List(host.Metadata.Namespace); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(list, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("list host succeed: %s", string(data))
		}
	}

	// 删除V1Host，也可使用 helper.V1.Host.Delete(host.Metadata.Namespace, host.Metadata.Name) 或 helper.V1.Host.DeleteHost(host.Metadata.Namespace, host.Metadata.Name)
	if h, err := registry.DeleteWithOpts(context.Background(), host.Metadata.Namespace, host.Metadata.Name, v1.WithSync); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("delete host succeed: %s", string(data))
		}
	}

	if h, err := registry.Get(host.Metadata.Namespace, host.Metadata.Name); err != nil {
		t.Error(err)
	} else if h != nil {
		t.Error("delete failed: object still exist")
	}
}
