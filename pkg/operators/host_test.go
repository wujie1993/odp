package operators_test

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

func TestAddHost(t *testing.T) {
	host := v1.NewV1Host()
	host.Metadata.Name = "host-32"
	host.Metadata.Annotations["ShortName"] = "节点32"
	host.Spec.SSH = v1.V1HostSSH{
		Host:     "172.25.21.32",
		Password: "admin",
		User:     "root",
		Port:     22,
	}

	helper := orm.GetHelper()

	if h, err := helper.V1.Host.Create(host); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("create host succeed: %s", string(data))
		}
	}
}
