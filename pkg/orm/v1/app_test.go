package v1_test

import (
	"context"
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

func TestAppEncoding(t *testing.T) {
	app := v1.NewV1App()

	bytes, err := json.MarshalIndent(app, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("encode: %s", string(bytes))

	if err := json.Unmarshal(bytes, app); err != nil {
		t.Fatal(err)
	}
	t.Logf("decode: %+v", *app)
}

func TestAppCRUD(t *testing.T) {
	app := v1.NewV1App()
	app.Metadata.Namespace = "default"
	app.Metadata.Name = "test_es_cluster"
	app.Spec.Category = core.AppCategoryThirdParty
	app.Spec.Versions = []v1.V1AppVersion{
		{
			Version:   "7.2.1",
			ShortName: "Elastisearch 集群",
			Desc:      "Elastisearch 7.2.1 集群",
			Platform:  core.AppPlatformBareMetal,
			Modules: []v1.V1AppModule{
				{
					Name: "es_cluster_7.2.1",
					HostLimits: v1.V1HostLimits{
						Max: 3,
						Min: 3,
					},
					Args: []v1.V1AppArgs{
						{
							Name:      "hosts",
							ShortName: "部署ES集群主机",
							Desc:      "多个地址用','分割",
							Type:      "IP:>=3",
						},
						{
							Name:      "es_data_path",
							ShortName: "ES数据目录",
							Type:      "String",
							Default:   "/opt/es/data",
						},
						{
							Name:      "es_log_path",
							ShortName: "ES日志目录",
							Type:      "String",
							Default:   "/opt/es/logs",
						},
					},
				},
			},
			SupportActions: []string{"install", "uninstall", "configure"},
		},
	}

	helper := orm.GetHelper()

	if h, err := helper.V1.App.Create(context.TODO(), app); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("create app succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.App.Get(context.TODO(), app.Metadata.Namespace, app.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("get app succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.App.UpdateStatusPhase(app.Metadata.Namespace, app.Metadata.Name, core.PhaseReady); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("update app succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.App.Get(context.TODO(), app.Metadata.Namespace, app.Metadata.Name); err != nil {
		t.Error(err)
	} else if h.(*v1.V1App).Status.Phase != core.PhaseReady {
		t.Error("update failed: result not equal to what you updated")
	}

	if list, err := helper.V1.App.List(context.TODO(), app.Metadata.Namespace); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(list, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("list app succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.App.Delete(context.TODO(), app.Metadata.Namespace, app.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("delete app succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.App.Get(context.TODO(), app.Metadata.Namespace, app.Metadata.Name); err != nil {
		t.Error(err)
	} else if h != nil {
		t.Error("delete failed: object still exist")
	}
}

func TestAppClean(t *testing.T) {
	helper := orm.GetHelper()
	list, err := helper.V1.App.List(context.TODO(), "default")
	if err != nil {
		t.Error(err)
	}

	for _, apiObj := range list {
		app := apiObj.(*v1.V1App)
		helper.V1.App.Delete(context.TODO(), app.Metadata.Namespace, app.Metadata.Name)
	}
}
