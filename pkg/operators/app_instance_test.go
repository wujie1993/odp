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
		//Endpoints: []string{"172.25.23.199:2379"},
	}
	db.InitKV()
}

func TestThridPartyInstall(t *testing.T) {
	appInstance := v1.NewV1AppInstance()
	appInstance.Metadata.Namespace = "default"
	appInstance.Metadata.Name = "test_mysql"
	appInstance.Metadata.Annotations["ShortName"] = "Mysql测试"
	appInstance.Spec.AppRef = v1.V1AppRef{
		Name:    "mysql",
		Version: "5.7.17",
	}
	appInstance.Spec.Modules = []v1.V1AppInstanceModule{
		v1.V1AppInstanceModule{
			Name:     "bare-mysql",
			HostRefs: []string{"host-172.25.21.32"},
			Args: []v1.V1AppInstanceArgs{
				v1.V1AppInstanceArgs{
					Name:  "mysql_datadir",
					Value: "/var/lib/mysql",
				},
				v1.V1AppInstanceArgs{
					Name:  "mysql_port",
					Value: 3306,
				},
				v1.V1AppInstanceArgs{
					Name:  "init",
					Value: false,
				},
			},
		},
	}
	appInstance.Spec.Action = core.AppActionInstall

	helper := orm.GetHelper()

	if h, err := helper.V1.AppInstance.Create(appInstance); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("create appInstance succeed: %s", string(data))
		}
	}
}

func TestAlgorithmInstanceInstall(t *testing.T) {
	appInstance := v1.NewV1AppInstance()
	appInstance.Metadata.Namespace = "default"
	appInstance.Metadata.Name = "videoanalysis-apps-85b3dba5"
	appInstance.Metadata.Annotations["ShortName"] = "视频解析测试"
	appInstance.Spec.AppRef = v1.V1AppRef{
		Name:    "videoanalysis-apps",
		Version: "v1.0.0-191",
	}
	appInstance.Spec.Modules = []v1.V1AppInstanceModule{
		v1.V1AppInstanceModule{
			Name:     "ice.ms",
			HostRefs: []string{"host-172.25.23.49"},
			Args: []v1.V1AppInstanceArgs{
				v1.V1AppInstanceArgs{
					Name:  "ALGORITHM_PLUGIN_NAME",
					Value: "HUMAN_EXTRA_JVIA_3.1.7.0_P4",
				},
				v1.V1AppInstanceArgs{
					Name:  "ALGORITHM_PLUGIN_VERSION",
					Value: "60404",
				},
				v1.V1AppInstanceArgs{
					Name:  "ALGORITHM_MEDIA_TYPE",
					Value: "image",
				},
				v1.V1AppInstanceArgs{
					Name:  "REQUEST_GPU",
					Value: true,
				},
			},
		},
	}
	appInstance.Spec.Global.Args = []v1.V1AppInstanceArgs{
		v1.V1AppInstanceArgs{
			Name:  "deploy_dir",
			Value: "/opt/ice.ms",
		},
	}
	appInstance.Spec.Action = core.AppActionInstall

	helper := orm.GetHelper()

	if h, err := helper.V1.AppInstance.Create(appInstance); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("create appInstance succeed: %s", string(data))
		}
	}
}
