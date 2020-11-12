package schedule_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"text/template"
	"time"

	"github.com/wujie1993/waves/pkg/ansible"
	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/schedule"
	"github.com/wujie1993/waves/pkg/setting"
)

func init() {
	setting.EtcdSetting = &setting.Etcd{
		Endpoints: []string{core.DefaultEtcdEndpoint},
	}
	db.InitKV()
}

func TestScheduler(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	s := schedule.NewScheduler("default")

	s.Run(ctx)
}

func TestValueJob(t *testing.T) {
	helper := orm.GetHelper()

	inventoryTpl, err := template.New("inventory").Parse(ansible.ANSIBLE_INVENTORY_HOST_INIT_TPL)
	if err != nil {
		t.Error(err)
	}
	var inventoryBuf bytes.Buffer
	host := v1.NewV1Host()
	host.Spec.SSH.Host = "172.25.23.199"
	host.Spec.SSH.User = "root"
	host.Spec.SSH.Password = "admin"
	host.Spec.SSH.Port = 22
	if err := inventoryTpl.Execute(&inventoryBuf, []*v1.V1Host{host}); err != nil {
		t.Error(err)
	}

	job := v1.NewV1Job()
	job.Metadata.Namespace = "default"
	job.Metadata.Name = "centos-init"
	job.Spec.Exec.Type = core.JobExecTypeAnsible
	job.Spec.Exec.Ansible.Bin = setting.AnsibleSetting.Bin
	job.Spec.Exec.Ansible.Inventories = []v1.V1AnsibleInventory{
		v1.V1AnsibleInventory{Value: ansible.ANSIBLE_INVENTORY_COMMON_TPL},
		v1.V1AnsibleInventory{Value: inventoryBuf.String()},
	}
	job.Spec.Exec.Ansible.Envs = []string{
		"act=install",
	}
	job.Spec.Exec.Ansible.Role = "centos-init"
	job.Spec.TimeoutSeconds = 60 * time.Second
	job.Spec.FailureThreshold = 3

	if _, err := helper.V1.Job.Create(job); err != nil {
		t.Error(err)
	}

	defer helper.V1.Job.Delete(job.Metadata.Namespace, job.Metadata.Name)

	ctx := context.Background()
	actionChan := helper.V1.Job.Watch(ctx, job.Metadata.Namespace, job.Metadata.Name)
	for action := range actionChan {
		job = action.Obj.(*v1.V1Job)
		phase := job.GetStatusPhase()
		t.Log(phase)
		switch phase {
		case core.PhaseCompleted:
			return
		case core.PhaseFailed:
			t.Error("job exec failed")
			return
		}
	}

	//time.Sleep(3 * time.Second)
}

func TestMultipleConfigMapJob(t *testing.T) {
	helper := orm.GetHelper()

	cmCommon := v1.NewV1ConfigMap()
	cmCommon.Metadata.Namespace = "default"
	cmCommon.Metadata.Name = fmt.Sprintf("centos-init-common-%d", time.Now().Unix())
	cmCommon.Spec["inventory"] = ansible.ANSIBLE_INVENTORY_COMMON_TPL
	if _, err := helper.V1.ConfigMap.Create(cmCommon); err != nil {
		t.Error(err)
	}
	defer helper.V1.ConfigMap.Delete(cmCommon.Metadata.Namespace, cmCommon.Metadata.Name)

	inventoryTpl, err := template.New("inventory").Parse(ansible.ANSIBLE_INVENTORY_HOST_INIT_TPL)
	if err != nil {
		t.Error(err)
	}
	var inventoryBuf bytes.Buffer
	host := v1.NewV1Host()
	host.Spec.SSH.Host = "172.25.23.199"
	host.Spec.SSH.User = "root"
	host.Spec.SSH.Password = "admin"
	host.Spec.SSH.Port = 22
	if err := inventoryTpl.Execute(&inventoryBuf, []*v1.V1Host{host}); err != nil {
		t.Error(err)
	}

	cmHost := v1.NewV1ConfigMap()
	cmHost.Metadata.Namespace = "default"
	cmHost.Metadata.Name = fmt.Sprintf("centos-init-inventory-%d", time.Now().Unix())
	cmHost.Spec["inventory"] = inventoryBuf.String()
	if _, err := helper.V1.ConfigMap.Create(cmHost); err != nil {
		t.Error(err)
	}
	defer helper.V1.ConfigMap.Delete(cmHost.Metadata.Namespace, cmHost.Metadata.Name)

	job := v1.NewV1Job()
	job.Metadata.Namespace = "default"
	job.Metadata.Name = "centos-init"
	job.Spec.Exec.Type = core.JobExecTypeAnsible
	job.Spec.Exec.Ansible.Bin = setting.AnsibleSetting.Bin
	job.Spec.Exec.Ansible.Inventories = []v1.V1AnsibleInventory{
		v1.V1AnsibleInventory{
			ValueFrom: core.ValueFrom{
				ConfigMapKeyRef: core.ConfigMapKeyRef{
					Name: cmCommon.Metadata.Name,
					Key:  "inventory",
				},
			},
		},
		v1.V1AnsibleInventory{
			ValueFrom: core.ValueFrom{
				ConfigMapKeyRef: core.ConfigMapKeyRef{
					Name: cmHost.Metadata.Name,
					Key:  "inventory",
				},
			},
		},
	}
	job.Spec.Exec.Ansible.Envs = []string{
		"act=install",
	}
	job.Spec.Exec.Ansible.Role = "centos-init"
	job.Spec.TimeoutSeconds = 60 * time.Second
	job.Spec.FailureThreshold = 1

	if _, err := helper.V1.Job.Create(job); err != nil {
		t.Error(err)
	}
	defer helper.V1.Job.Delete(job.Metadata.Namespace, job.Metadata.Name)

	ctx := context.Background()
	actionChan := helper.V1.Job.Watch(ctx, job.Metadata.Namespace, job.Metadata.Name)
	for action := range actionChan {
		job = action.Obj.(*v1.V1Job)
		phase := job.GetStatusPhase()
		t.Log(phase)
		if phase == core.PhaseCompleted {
			break
		}
	}

	//time.Sleep(3 * time.Second)
}

func TestSingleConfigMapJob(t *testing.T) {
	helper := orm.GetHelper()

	// 模板配置解析
	inventoryTpl, err := template.New("inventory").Parse(ansible.ANSIBLE_INVENTORY_HOST_INIT_TPL)
	if err != nil {
		t.Error(err)
	}
	var inventoryBuf bytes.Buffer
	host := v1.NewV1Host()
	host.Spec.SSH.Host = "172.25.23.199"
	host.Spec.SSH.User = "root"
	host.Spec.SSH.Password = "admin"
	host.Spec.SSH.Port = 22
	if err := inventoryTpl.Execute(&inventoryBuf, []*v1.V1Host{host}); err != nil {
		t.Error(err)
	}

	// 创建ConfigMap
	configMap := v1.NewV1ConfigMap()
	configMap.Metadata.Namespace = "default"
	configMap.Metadata.Name = fmt.Sprintf("centos-init-%d", time.Now().Unix())
	configMap.Spec["common"] = ansible.ANSIBLE_INVENTORY_COMMON_TPL
	configMap.Spec["inventory"] = inventoryBuf.String()
	if _, err := helper.V1.ConfigMap.Create(configMap); err != nil {
		t.Error(err)
	}
	defer helper.V1.ConfigMap.Delete(configMap.Metadata.Namespace, configMap.Metadata.Name)

	job := v1.NewV1Job()
	job.Metadata.Namespace = "default"
	job.Metadata.Name = "centos-init"
	job.Spec.Exec.Type = core.JobExecTypeAnsible
	job.Spec.Exec.Ansible.Bin = setting.AnsibleSetting.Bin
	job.Spec.Exec.Ansible.Inventories = []v1.V1AnsibleInventory{
		v1.V1AnsibleInventory{
			ValueFrom: core.ValueFrom{
				ConfigMapKeyRef: core.ConfigMapKeyRef{
					Name: configMap.Metadata.Name,
				},
			},
		},
	}
	job.Spec.Exec.Ansible.Envs = []string{
		"act=install",
	}
	job.Spec.Exec.Ansible.Role = "centos-init"
	job.Spec.TimeoutSeconds = 60 * time.Second
	job.Spec.FailureThreshold = 1

	if _, err := helper.V1.Job.Create(job); err != nil {
		t.Error(err)
	}
	defer helper.V1.Job.Delete(job.Metadata.Namespace, job.Metadata.Name)

	ctx := context.Background()
	actionChan := helper.V1.Job.Watch(ctx, job.Metadata.Namespace, job.Metadata.Name)
	for action := range actionChan {
		job = action.Obj.(*v1.V1Job)
		phase := job.GetStatusPhase()
		t.Log(phase)
		if phase == core.PhaseCompleted {
			break
		}
	}

	//time.Sleep(3 * time.Second)
}
