package v1_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
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

func TestAppInstanceEncoding(t *testing.T) {
	appInstance := v1.NewV1AppInstance()

	bytes, err := json.MarshalIndent(appInstance, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("encode: %s", string(bytes))

	if err := json.Unmarshal(bytes, appInstance); err != nil {
		t.Fatal(err)
	}
	t.Logf("decode: %+v", *appInstance)
}

func TestAppInstanceCRUD(t *testing.T) {
	helper := orm.GetHelper()

	appInstance := v1.NewV1AppInstance()
	appInstance.Metadata.Namespace = "default"
	appInstance.Metadata.Name = "test_es_cluster_0"
	appInstance.Metadata.Annotations["shortName"] = "我的ES集群"
	appInstance.Metadata.Annotations["desc"] = "Elastisearch 7.2.1 高可用集群"
	appInstance.Spec.AppRef = v1.V1AppRef{
		Name:    "es_cluster",
		Version: "7.2.1",
	}
	appInstance.Spec.Modules = []v1.V1AppInstanceModule{
		{
			HostRefs: []string{"node-25", "node-32", "node-99"},
			Args: []v1.V1AppInstanceArgs{
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
	}
	appInstance.Spec.Action = core.AppActionInstall

	/*
		if h, err := helper.V1.AppInstance.Create(context.TODO(), appInstance); err != nil {
			t.Error(err)
		} else {
			if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
				t.Error(err)
			} else {
				t.Logf("create appInstance succeed: %s", string(data))
			}
		}
	*/

	if h, err := helper.V1.AppInstance.Get(context.TODO(), appInstance.Metadata.Namespace, appInstance.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("get appInstance succeed: %s", string(data))
		}
	}

	/*
		appInstance.Spec.Action = core.AppActionUninstall
		appInstance.Status.Phase = core.PhaseReady
		if h, err := helper.V1.AppInstance.Update(context.TODO(), appInstance, core.WithStatus()); err != nil {
			t.Error(err)
		} else {
			if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
				t.Error(err)
			} else {
				t.Logf("update appInstance succeed: %s", string(data))
			}
		}

		if h, err := helper.V1.AppInstance.Get(context.TODO(), appInstance.Metadata.Namespace, appInstance.Metadata.Name); err != nil {
			t.Error(err)
		} else if h.(*v1.V1AppInstance).Status.Phase != core.PhaseReady || h.(*v1.V1AppInstance).Spec.Action != core.AppActionUninstall {
			t.Error("update failed: result not equal to what you updated")
		}

		if list, err := helper.V1.AppInstance.List(context.TODO(), appInstance.Metadata.Namespace); err != nil {
			t.Error(err)
		} else {
			if data, err := json.MarshalIndent(list, "", "\t"); err != nil {
				t.Fatal(err)
			} else {
				t.Logf("list appInstance succeed: %s", string(data))
			}
		}

		if h, err := helper.V1.AppInstance.Delete(context.TODO(), appInstance.Metadata.Namespace, appInstance.Metadata.Name); err != nil {
			t.Error(err)
		} else {
			if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
				t.Error(err)
			} else {
				t.Logf("delete appInstance succeed: %s", string(data))
			}
		}

		if h, err := helper.V1.AppInstance.Get(context.TODO(), appInstance.Metadata.Namespace, appInstance.Metadata.Name); err != nil {
			t.Error(err)
		} else if h != nil {
			t.Error("delete failed: object still exist")
		}
	*/
}

/*
func TestAppInstanceClean(t *testing.T) {
	helper := orm.GetHelper()
	list, err := helper.V1.AppInstance.List("default")
	if err != nil {
		t.Error(err)
	}

	for _, apiObj := range list {
		appInstance := apiObj.(*v1.V1AppInstance)
		helper.V1.AppInstance.Delete(appInstance.Metadata.Namespace, appInstance.Metadata.Name)
	}
}
*/
