package v1_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/wujie1993/waves/pkg/ansible"
	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/setting"
)

func init() {
	setting.EtcdSetting = &setting.Etcd{
		Endpoints: []string{core.DefaultEtcdEndpoint},
		// Endpoints: []string{"localhost:2378"},
	}
	setting.AnsibleSetting = &setting.Ansible{
		BaseDir: "/root/Projects/pcitech/devops/base_dir",
		TplsDir: "../../../conf/tpls",
	}
	db.InitKV()
}

func TestJobEncoding(t *testing.T) {
	job := v1.NewV1Job()

	bytes, err := json.MarshalIndent(job, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("encode: %s", string(bytes))

	if err := json.Unmarshal(bytes, job); err != nil {
		t.Fatal(err)
	}
	t.Logf("decode: %+v", *job)
}

func TestJobCRUD(t *testing.T) {
	helper := orm.GetHelper()

	commonInventoryStr, err := ansible.RenderCommonInventory()
	if err != nil {
		t.Error(err)
		return
	}

	job := v1.NewV1Job()
	job.Metadata.Namespace = "default"
	job.Metadata.Name = "my_es"
	job.Metadata.Annotations["cluster"] = "my_cluster"
	job.Spec.Exec.Type = core.JobExecTypeAnsible
	job.Spec.Exec.Ansible.Bin = setting.AnsibleSetting.Bin
	job.Spec.Exec.Ansible.Inventories = []v1.V1AnsibleInventory{
		{
			Value: commonInventoryStr,
		},
	}
	job.Spec.Exec.Ansible.Envs = []string{
		"act=install",
	}
	job.Spec.Exec.Ansible.Playbook = "playbooks/gen_es_cluster-7.2.1.yml"

	if h, err := helper.V1.Job.Create(job); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("create job succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.Job.Get(job.Metadata.Namespace, job.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("get job succeed: %s", string(data))
		}
	}

	job.Spec.Exec.Ansible.Bin = "/usr/local/bin/ansible-playbook"
	if h, err := helper.V1.Job.Update(job, false); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("update job succeed: %s", string(data))
		}
	}

	if obj, err := helper.V1.Job.Get(job.Metadata.Namespace, job.Metadata.Name); err != nil {
		t.Error(err)
	} else if obj.(*v1.V1Job).Spec.Exec.Ansible.Bin != "/usr/local/bin/ansible-playbook" {
		t.Error("update failed: result not equal to what you updated")
	}

	if list, err := helper.V1.Job.List(job.Metadata.Namespace); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(list, "", "\t"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("list job succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.Job.Delete(job.Metadata.Namespace, job.Metadata.Name); err != nil {
		t.Error(err)
	} else {
		if data, err := json.MarshalIndent(h, "", "\t"); err != nil {
			t.Error(err)
		} else {
			t.Logf("delete job succeed: %s", string(data))
		}
	}

	if h, err := helper.V1.Job.Get(job.Metadata.Namespace, job.Metadata.Name); err != nil {
		t.Error(err)
	} else if h != nil {
		t.Error("delete failed: object still exist")
	}
}

func TestJobWatch(t *testing.T) {
	helper := orm.GetHelper()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	actionWatcher := helper.V1.Job.Watch(ctx, "default", "")
	go func() {
		for action := range actionWatcher {
			if data, err := json.MarshalIndent(action.Obj, "", "\t"); err != nil {
				t.Error(err)
			} else {
				t.Logf("create job succeed: %s", string(data))
			}
		}
	}()
	TestJobCRUD(t)

	time.Sleep(3 * time.Second)
}

func TestJobGetLog(t *testing.T) {
	helper := orm.GetHelper()

	data, err := helper.V1.Job.GetLog("/root/Projects/pcitech/devops/web-deploy/jobs", "mysql-d9594aa7--1588235423")
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(string(data))
}

/*
func TestJobWatchLog(t *testing.T) {
	helper := orm.GetHelper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	logWatcher, err := helper.V1.Job.WatchLog(ctx, "/root/Projects/pcitech/devops/web-deploy/jobs", "0bf4c157-3875-48f2-ab14-41d7420dacf7")
	if err != nil {
		t.Error(err)
		return
	}

	for {
		select {
		case line, ok := <-logWatcher:
			if !ok {
				return
			}
			t.Log(line)
		}
	}
}
*/
