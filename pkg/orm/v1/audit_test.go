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

func TestAuditEncoding(t *testing.T) {
	audit := v1.NewV1Audit()

	bytes, err := json.MarshalIndent(audit, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("encode: %s", string(bytes))

	if err := json.Unmarshal(bytes, audit); err != nil {
		t.Fatal(err)
	}
	t.Logf("decode: %+v", *audit)
}

func TestAuditCRUD(t *testing.T) {
	audit := v1.NewV1Audit()
	audit.Metadata.Name = "1587092917"
	audit.Spec.ResourceRef = v1.V1ResourceRef{
		Kind:      core.KindAppInstance,
		Name:      "my_es",
		Namespace: core.DefaultNamespace,
	}
	audit.Spec.Action = core.AuditActionCreate
	audit.Spec.Msg = "用户 xxxx 在命名空间 default 下创建了应用实例 my_es"
	audit.Spec.SourceIP = "172.21.25.50"

	helper := orm.GetHelper()

	if h, err := helper.V1.Audit.Create(audit); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("create audit succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.Audit.Get(audit.Metadata.Namespace, audit.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("get audit succeed: %s", string(data))
		}
	}

	audit.Spec.Action = core.AuditActionUpdate
	if h, err := helper.V1.Audit.Update(audit, false); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("update audit succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.Audit.Get(audit.Metadata.Namespace, audit.Metadata.Name); err != nil {
		t.Error(err)
	} else if h.(*v1.V1Audit).Spec.Action != core.AuditActionUpdate {
		t.Error("update failed: result not equal to what you updated")
	}

	if list, err := helper.V1.Audit.List(audit.Metadata.Namespace); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(list, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("list audit succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.Audit.Delete(audit.Metadata.Namespace, audit.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("delete audit succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.Audit.Get(audit.Metadata.Namespace, audit.Metadata.Name); err != nil {
		t.Error(err)
	} else if h != nil {
		t.Error("delete failed: object still exist")
	}
}
