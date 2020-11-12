package operators

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

// EventOperator 事件控制器
type EventOperator struct {
	BaseOperator
}

func (o *EventOperator) handleEvent(ctx context.Context, obj core.ApiObject) {
	event := obj.(*v1.Event)
	log.Tracef("%s '%s' is %s", event.Kind, event.GetKey(), event.Status.Phase)

	switch event.Status.Phase {
	case core.PhaseDeleting:
		// 如果资源正在删除中，则跳过
		if _, ok := o.deletings.Get(event.GetKey()); ok {
			return
		}
		o.deletings.Set(event.GetKey(), event.SpecHash())
		defer o.deletings.Unset(event.GetKey())

		if len(event.Metadata.Finalizers) > 0 {
			// 每次只处理一项Finalizer
			switch event.Metadata.Finalizers[0] {
			case core.FinalizerCleanRefJob:
				// 同步删除关联的任务
				if event.Spec.JobRef != "" {
					if _, err := o.helper.V1.Job.Delete(context.TODO(), "", event.Spec.JobRef, core.WithSync()); err != nil {
						log.Error(err)
						return
					}
				}
			}

			o.deletings.Unset(event.GetKey())
			event.Metadata.Finalizers = event.Metadata.Finalizers[1:]
			if _, err := o.helper.V1.Event.Update(context.TODO(), event, core.WithFinalizer()); err != nil {
				log.Error(err)
				return
			}
		} else {
			if _, err := o.helper.V1.Event.Delete(context.TODO(), "", event.Metadata.Name); err != nil {
				log.Error(err)
				return
			}
		}
	}
}

func NewEventOperator() *EventOperator {
	o := &EventOperator{
		BaseOperator: NewBaseOperator(v1.NewEventRegistry()),
	}
	o.SetHandleFunc(o.handleEvent)
	return o
}
