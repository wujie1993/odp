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

func TestEventEncoding(t *testing.T) {
	event := v1.NewV1Event()

	bytes, err := json.MarshalIndent(event, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("encode: %s", string(bytes))

	if err := json.Unmarshal(bytes, event); err != nil {
		t.Fatal(err)
	}
	t.Logf("decode: %+v", *event)
}

func TestEventCRUD(t *testing.T) {
	event := v1.NewV1Event()
	event.Metadata.Name = "1587092917"
	event.Spec.ResourceRef = v1.V1ResourceRef{
		Kind:      core.KindEvent,
		Name:      "my_es",
		Namespace: core.DefaultNamespace,
	}
	event.Spec.Action = core.EventActionInstall
	event.Spec.Msg = "用户 xxxx 在命名空间 default 下安装了应用实例 my_es"
	event.Spec.JobRef = ""

	helper := orm.GetHelper()

	if h, err := helper.V1.Event.Create(event); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("create event succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.Event.Get(event.Metadata.Namespace, event.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("get event succeed: %s", string(data))
		}
	}

	event.Spec.Action = core.EventActionConfigure
	event.Spec.Msg = "用户 xxxx 在命名空间 default 下配置了应用实例 my_es"
	if h, err := helper.V1.Event.Update(event, false); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("update event succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.Event.Get(event.Metadata.Namespace, event.Metadata.Name); err != nil {
		t.Error(err)
	} else if h.(*v1.V1Event).Spec.Action != core.EventActionConfigure {
		t.Error("update failed: result not equal to what you updated")
	}

	if list, err := helper.V1.Event.List(event.Metadata.Namespace); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(list, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("list event succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.Event.Delete(event.Metadata.Namespace, event.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("delete event succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.Event.Get(event.Metadata.Namespace, event.Metadata.Name); err != nil {
		t.Error(err)
	} else if h != nil {
		t.Error("delete failed: object still exist")
	}
}

func TestEventRecord(t *testing.T) {
	helper := orm.GetHelper()

	// 事件开始
	e1 := v1.NewV1Event()
	e1.Spec.ResourceRef.Kind = core.KindHost
	e1.Spec.ResourceRef.Name = "host-192.168.1.1"
	e1.Spec.Action = core.EventActionInitial
	e1.Spec.JobRef = "job-host-192.168.1.1-init"
	if err := helper.V1.Event.Record(e1); err != nil {
		t.Error(err)
	}

	// 事件完成
	e2 := v1.NewV1Event()
	e2.Spec.ResourceRef.Kind = core.KindHost
	e2.Spec.ResourceRef.Name = "host-192.168.1.1"
	e2.Spec.Action = core.EventActionInitial
	e2.Spec.JobRef = "job-host-192.168.1.1-init"
	e2.Status.Phase = core.PhaseCompleted
	if err := helper.V1.Event.Record(e2); err != nil {
		t.Error(err)
	}

	// 事件失败
	e3 := v1.NewV1Event()
	e3.Spec.ResourceRef.Kind = core.KindHost
	e3.Spec.ResourceRef.Name = "host-192.168.1.1"
	e3.Spec.Action = core.EventActionInitial
	e3.Spec.JobRef = "job-host-192.168.1.1-init"
	e3.Status.Phase = core.PhaseFailed
	if err := helper.V1.Event.Record(e3); err != nil {
		t.Error(err)
	}
}

/*
func TestEventClean(t *testing.T) {
	helper := orm.GetHelper()
	list, err := helper.V1.Event.List("")
	if err != nil {
		t.Error(err)
	}

	for _, apiObj := range list {
		event := apiObj.(*v1.V1Event)
		helper.V1.Event.Delete("", event.Metadata.Name)
	}
}
*/
