package orm_test

import (
	"encoding/json"
	"testing"

	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/orm/v2"
)

func TestConvert(t *testing.T) {
	appInstance := v1.NewAppInstance()
	appInstance.Metadata.Namespace = "default"
	appInstance.Metadata.Name = "test_es_cluster"
	appInstance.Metadata.Annotations["shortName"] = "我的ES集群"
	appInstance.Metadata.Annotations["desc"] = "Elastisearch 7.2.1 高可用集群"
	appInstance.Spec.AppRef = v1.AppRef{
		Name:    "test_es_cluster",
		Version: "7.2.1",
	}
	appInstance.Spec.Modules = []v1.AppInstanceModule{
		v1.AppInstanceModule{
			HostRefs: []string{"node-25", "node-32", "node-99"},
			Args: []v1.AppInstanceArgs{
				v1.AppInstanceArgs{
					Name:  "es_data_path",
					Value: "/opt/es/data",
				},
				v1.AppInstanceArgs{
					Name:  "es_log_path",
					Value: "/opt/es/logs",
				},
			},
		},
	}
	appInstance.Spec.Action = core.AppActionInstall

	rtObj, err := orm.Convert(appInstance, core.GVK{Group: core.Group, ApiVersion: v2.ApiVersion, Kind: core.KindAppInstance})
	if err != nil {
		t.Fatal(err)
	}

	bytes, err := json.MarshalIndent(rtObj, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("encode: %s", string(bytes))
}
