package v1_test

import (
	"encoding/json"
	"testing"

	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/orm/v1"

	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/setting"
)

func init() {
	setting.EtcdSetting = &setting.Etcd{
		Endpoints: []string{core.DefaultEtcdEndpoint},
	}
	db.InitKV()
}

func TestGPUEncoding(t *testing.T) {
	gpu := v1.NewV1GPU()

	bytes, err := json.MarshalIndent(gpu, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("encode: %s", string(bytes))

	if err := json.Unmarshal(bytes, gpu); err != nil {
		t.Fatal(err)
	}
	t.Logf("decode: %+v", *gpu)
}

func TestGPUCRUD(t *testing.T) {
	gpu := v1.NewV1GPU()
	gpu.Metadata.Namespace = "test-gpu"

	helper := orm.GetHelper()

	if h, err := helper.V1.GPU.Create(gpu); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("create gpu succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.GPU.Get(gpu.Metadata.Namespace, gpu.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("get gpu succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.GPU.UpdateStatusPhase(gpu.Metadata.Namespace, gpu.Metadata.Name, core.PhaseReady); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("update gpu succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.GPU.Get(gpu.Metadata.Namespace, gpu.Metadata.Name); err != nil {
		t.Error(err)
	} else if h.(*v1.V1GPU).Status.Phase != core.PhaseReady {
		t.Error("update failed: result not equal to what you updated")
	}

	if list, err := helper.V1.GPU.List(gpu.Metadata.Namespace); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(list, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("list gpu succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.GPU.Delete(gpu.Metadata.Namespace, gpu.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("delete gpu succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.GPU.Get(gpu.Metadata.Namespace, gpu.Metadata.Name); err != nil {
		t.Error(err)
	} else if h != nil {
		t.Error("delete failed: object still exist")
	}
}

func TestGPUClean(t *testing.T) {
	helper := orm.GetHelper()
	list, err := helper.V1.GPU.List("default")
	if err != nil {
		t.Error(err)
	}

	for _, apiObj := range list {
		gpu := apiObj.(*v1.V1GPU)
		helper.V1.GPU.Delete(gpu.Metadata.Namespace, gpu.Metadata.Name)
	}
}

func TestUnbindAllGPU(t *testing.T) {
	helper := orm.GetHelper()
	list, err := helper.V1.GPU.List("default")
	if err != nil {
		t.Error(err)
	}

	for _, apiObj := range list {
		gpu := apiObj.(*v1.V1GPU)
		gpu.Spec.AppInstanceModuleRef = v1.V1AppInstanceModuleRef{}
		gpu.Status.Phase = core.PhaseWaiting
		if _, err := helper.V1.GPU.Update(gpu, true); err != nil {
			t.Error(err)
		}
	}
}
