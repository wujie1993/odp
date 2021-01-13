package operators

import (
	"context"
	"os"
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v2"
	"github.com/wujie1993/waves/pkg/setting"
)

// JobOperator 任务控制器
type JobOperator struct {
	BaseOperator
}

func (o *JobOperator) handleJob(ctx context.Context, obj core.ApiObject) error {
	job := obj.(*v2.Job)
	log.Tracef("%s '%s' is %s", job.Kind, job.GetKey(), job.Status.Phase)

	switch job.Status.Phase {
	case core.PhaseDeleting:
		return o.handleDeleting(ctx, obj)
	}
	return nil
}

func (o JobOperator) finalizeJob(ctx context.Context, obj core.ApiObject) error {
	job := obj.(*v2.Job)

	// 每次只处理一项Finalizer
	switch job.Metadata.Finalizers[0] {
	case core.FinalizerCleanRefConfigMap:
		for _, play := range job.Spec.Exec.Ansible.Plays {
			// 删除关联的inventory
			if play.Inventory.ValueFrom.ConfigMapRef.Name != "" {
				if _, err := o.helper.V1.ConfigMap.Delete(context.TODO(), play.Inventory.ValueFrom.ConfigMapRef.Namespace, play.Inventory.ValueFrom.ConfigMapRef.Name, core.WithSync()); err != nil {
					log.Error(err)
					return err
				}
			}
			// 删除关联的group_vars
			if play.GroupVars.ValueFrom.ConfigMapRef.Name != "" {
				if _, err := o.helper.V1.ConfigMap.Delete(context.TODO(),
					play.GroupVars.ValueFrom.ConfigMapRef.Namespace,
					play.GroupVars.ValueFrom.ConfigMapRef.Name, core.WithSync()); err != nil {
					log.Error(err)
					return err
				}
			}
		}
	case core.FinalizerCleanJobWorkDir:
		if err := os.RemoveAll(path.Join(setting.AppSetting.DataDir, setting.JobsDir, job.GetMetadata().Uid)); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

func NewJobOperator() *JobOperator {
	o := &JobOperator{
		BaseOperator: NewBaseOperator(v2.NewJobRegistry()),
	}
	o.SetHandleFunc(o.handleJob)
	o.SetFinalizeFunc(o.finalizeJob)
	return o
}
