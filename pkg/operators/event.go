package operators

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

// EventOperator 事件管理器
type EventOperator struct {
	BaseOperator
}

// handleEvent 处理事件的变更操作
func (o *EventOperator) handleEvent(ctx context.Context, obj core.ApiObject) error {
	event := obj.(*v1.Event)
	log.Tracef("%s '%s' is %s", event.Kind, event.GetKey(), event.Status.Phase)

	switch event.Status.Phase {
	case core.PhaseDeleting:
		o.delete(ctx, obj)
	}
	return nil
}

// finalizeEvent 级联清除事件的关联资源
func (o EventOperator) finalizeEvent(ctx context.Context, obj core.ApiObject) error {
	event := obj.(*v1.Event)

	// 每次只处理一项Finalizer
	switch event.Metadata.Finalizers[0] {
	case core.FinalizerCleanRefJob:
		// 同步删除关联的任务
		if event.Spec.JobRef != "" {
			if _, err := o.helper.V1.Job.Delete(context.TODO(), "", event.Spec.JobRef); err != nil {
				log.Error(err)
				return err
			}
		}
	}
	return nil
}

// NewEventOperator 创建事件管理器
func NewEventOperator() *EventOperator {
	o := &EventOperator{
		BaseOperator: NewBaseOperator(v1.NewEventRegistry()),
	}
	o.SetHandleFunc(o.handleEvent)
	o.SetFinalizeFunc(o.finalizeEvent)
	return o
}
