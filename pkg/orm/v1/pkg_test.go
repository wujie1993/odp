package v1_test

import (
	"encoding/json"
	"testing"

	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/setting"
)

func init() {
	setting.EtcdSetting = &setting.Etcd{
		Endpoints: []string{core.DefaultEtcdEndpoint},
	}
	db.InitKV()
}

func TestPkgEncoding(t *testing.T) {
	pkg := v1.NewV1Pkg()

	bytes, err := json.MarshalIndent(pkg, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("encode: %s", string(bytes))

	if err := json.Unmarshal(bytes, pkg); err != nil {
		t.Fatal(err)
	}
	t.Logf("decode: %+v", *pkg)
}

func TestPkgCRUD(t *testing.T) {
	pkg := v1.NewV1Pkg()
	pkg.Metadata.Name = "20200410_14500203_nacos_server_v1.0.2_22020020-2"
	pkg.Spec.Author = "anoymous@pcitech.com"
	pkg.Spec.Desc = "轻配置中心"
	pkg.Spec.Images = []string{"harbor.pcitech.com/base/nacos-server:1.1.4"}
	pkg.Spec.Module = "nacos-server"
	pkg.Spec.Version = "1.1.4"
	pkg.Spec.Platform = core.AppPlatformK8s
	pkg.Spec.Provision = core.PkgProvisionFull
	pkg.Spec.Synced = false

	helper := orm.GetHelper()

	if h, err := helper.V1.Pkg.Create(pkg); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("create pkg succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.Pkg.Get(pkg.Metadata.Namespace, pkg.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("get pkg succeed: %s", string(data))
		}
	}

	pkg.Spec.Synced = true
	if h, err := helper.V1.Pkg.Update(pkg, false); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("update pkg succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.Pkg.Get(pkg.Metadata.Namespace, pkg.Metadata.Name); err != nil {
		t.Error(err)
	} else if h.(*v1.V1Pkg).Spec.Synced != true {
		t.Error("update failed: result not equal to what you updated")
	}

	if list, err := helper.V1.Pkg.List(pkg.Metadata.Namespace); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(list, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("list pkg succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.Pkg.Delete(pkg.Metadata.Namespace, pkg.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("delete pkg succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.Pkg.Get(pkg.Metadata.Namespace, pkg.Metadata.Name); err != nil {
		t.Error(err)
	} else if h != nil {
		t.Error("delete failed: object still exist")
	}
}
