package operators

import (
	"context"
	"os"
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/setting"
)

// JobOperator 任务控制器
type JobOperator struct {
	BaseOperator
}

func (o *JobOperator) handleJob(ctx context.Context, obj core.ApiObject) {
	job := obj.(*v1.Job)
	log.Tracef("%s '%s' is %s", job.Kind, job.GetKey(), job.Status.Phase)

	switch job.Status.Phase {
	case core.PhaseDeleting:
		// 如果资源正在删除中，则跳过
		if _, ok := o.deletings.Get(job.GetKey()); ok {
			return
		}
		o.deletings.Set(job.GetKey(), job.SpecHash())
		defer o.deletings.Unset(job.GetKey())

		if len(job.Metadata.Finalizers) > 0 {
			// 每次只处理一项Finalizer
			switch job.Metadata.Finalizers[0] {
			case core.FinalizerCleanRefConfigMap:
				// 删除关联的inventory
				for _, inventory := range job.Spec.Exec.Ansible.Inventories {
					if inventory.ValueFrom.ConfigMapRef.Name != "" {
						if _, err := o.helper.V1.ConfigMap.Delete(context.TODO(), inventory.ValueFrom.ConfigMapRef.Namespace, inventory.ValueFrom.ConfigMapRef.Name, core.WithSync()); err != nil {
							log.Error(err)
							return
						}
					}
				}
				// 删除关联的group_vars
				if job.Spec.Exec.Ansible.GroupVars.ValueFrom.ConfigMapRef.Name != "" {
					if _, err := o.helper.V1.ConfigMap.Delete(context.TODO(),
						job.Spec.Exec.Ansible.GroupVars.ValueFrom.ConfigMapRef.Namespace,
						job.Spec.Exec.Ansible.GroupVars.ValueFrom.ConfigMapRef.Name, core.WithSync()); err != nil {
						log.Error(err)
						return
					}
				}
			case core.FinalizerCleanJobWorkDir:
				if err := os.RemoveAll(path.Join(setting.AppSetting.DataDir, setting.JobsDir, job.GetMetadata().Uid)); err != nil {
					log.Error(err)
					return
				}
			}

			o.deletings.Unset(job.GetKey())
			job.Metadata.Finalizers = job.Metadata.Finalizers[1:]
			if _, err := o.helper.V1.Job.Update(context.TODO(), job, core.WithFinalizer()); err != nil {
				log.Error(err)
				return
			}
		} else {
			if _, err := o.helper.V1.Job.Delete(context.TODO(), "", job.Metadata.Name); err != nil {
				log.Error(err)
				return
			}
		}
	}
}

func NewJobOperator() *JobOperator {
	o := &JobOperator{
		BaseOperator: NewBaseOperator(v1.NewJobRegistry()),
	}
	o.SetHandleFunc(o.handleJob)
	return o
}
