package v2_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v2"
	"github.com/wujie1993/waves/pkg/setting"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	log.SetReportCaller(true)

	setting.EtcdSetting = &setting.Etcd{
		Endpoints: []string{core.DefaultEtcdEndpoint},
	}
	db.InitKV()
}

func newAppInstanceExample() *v2.AppInstance {
	appInstance := v2.NewAppInstance()
	appInstance.Metadata.Namespace = "default"
	appInstance.Metadata.Name = "test_es_cluster_0"
	appInstance.Metadata.Annotations["shortName"] = "我的ES集群"
	appInstance.Metadata.Annotations["desc"] = "Elastisearch 7.2.1 高可用集群"
	appInstance.Spec.AppRef = v2.AppRef{
		Name:    "es_cluster",
		Version: "7.2.1",
	}
	appInstance.Spec.Modules = []v2.AppInstanceModule{
		{
			Name: "test_es_cluster",
			Replicas: []v2.AppInstanceModuleReplica{
				{
					HostRefs: []string{"node-25", "node-32", "node-99"},
					Args: []v2.AppInstanceArgs{
						{
							Name:  "es_data_path",
							Value: "/opt/es/data",
						},
						{
							Name:  "es_log_path",
							Value: "/opt/es/logs",
						},
					},
				},
			},
		},
	}
	appInstance.Spec.Action = core.AppActionInstall
	appInstance.Spec.Global.Args = []v2.AppInstanceArgs{}
	return appInstance
}

func TestAppInstanceJSON(t *testing.T) {
	src := newAppInstanceExample()

	data, err := src.ToJSONPretty()
	if err != nil {
		t.Fatal(err)
	}

	dst := new(v2.AppInstance)
	if err := dst.FromJSON(data); err != nil {
		t.Fatal(err)
	}

	srcHash := src.Sha256()
	dstHash := dst.Sha256()
	if srcHash != dstHash {
		t.Errorf("json encode result not equal. src: %s dst %s", srcHash, dstHash)
	}
}

func TestAppInstanceYAML(t *testing.T) {
	src := newAppInstanceExample()

	data, err := src.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	dst := new(v2.AppInstance)
	if err := dst.FromYAML(data); err != nil {
		t.Fatal(err)
	}

	srcHash := src.Sha256()
	dstHash := dst.Sha256()
	if srcHash != dstHash {
		t.Errorf("json encode result not equal. src: %s dst %s", srcHash, dstHash)
	}
}

func TestAppInstanceDeepCopy(t *testing.T) {
	src := newAppInstanceExample()

	dst := src.DeepCopy()
	if src.Sha256() != dst.Sha256() {
		t.Errorf("deepcopy result not equal")
	}
}

func TestAppInstanceCRUD(t *testing.T) {
	helper := orm.GetHelper()

	appInstance := newAppInstanceExample()

	if h, err := helper.V2.AppInstance.Create(context.TODO(), appInstance); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("create appInstance succeed: %s", string(data))
		}
	}

	if h, err := helper.V2.AppInstance.Get(context.TODO(), appInstance.Metadata.Namespace, appInstance.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("get appInstance succeed: %s", string(data))
		}
	}
	appInstance.Spec.Action = core.AppActionUninstall
	appInstance.Status.Phase = core.PhaseReady
	if h, err := helper.V2.AppInstance.Update(context.TODO(), appInstance, core.WithStatus()); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("update appInstance succeed: %s", string(data))
		}
	}

	if h, err := helper.V2.AppInstance.Get(context.TODO(), appInstance.Metadata.Namespace, appInstance.Metadata.Name); err != nil {
		t.Error(err)
	} else if h == nil {
		t.Fatalf("%s not found", appInstance.GetKey())
	} else if h.(*v2.AppInstance).Status.Phase != core.PhaseReady || h.(*v2.AppInstance).Spec.Action != core.AppActionUninstall {
		t.Error("update failed: result not equal to what you updated")
	}

	if list, err := helper.V2.AppInstance.List(context.TODO(), appInstance.Metadata.Namespace); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(list, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("list appInstance succeed: %s", string(data))
		}
	}

	if h, err := helper.V2.AppInstance.Delete(context.TODO(), appInstance.Metadata.Namespace, appInstance.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("delete appInstance succeed: %s", string(data))
		}
	}

	if h, err := helper.V2.AppInstance.Get(context.TODO(), appInstance.Metadata.Namespace, appInstance.Metadata.Name); err != nil {
		t.Error(err)
	} else if h != nil {
		t.Error("delete failed: object still exist")
	}
}

func TestMigrateAppInstance(t *testing.T) {
	registry := v2.NewAppInstanceRegistry()
	if err := registry.MigrateObjects(); err != nil {
		t.Fatal(err)
	}
}
