package operators

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

// AppOperator 应用控制器
type AppOperator struct {
	BaseOperator
}

func (c *AppOperator) handleApp(ctx context.Context, obj core.ApiObject) {
	app := obj.(*v1.App)
	log.Tracef("%s '%s' is %s", app.Kind, app.GetKey(), app.Status.Phase)

	switch app.Status.Phase {
	case core.PhaseDeleting:
		// 如果资源正在删除中，则跳过
		if _, ok := c.deletings.Get(app.GetKey()); ok {
			return
		}
		c.deletings.Set(app.GetKey(), app.SpecHash())
		defer c.deletings.Unset(app.GetKey())

		if len(app.Metadata.Finalizers) > 0 {
			// 每次只处理一项Finalizer
			switch app.Metadata.Finalizers[0] {
			case core.FinalizerCleanRefConfigMap:
				// 同步删除关联的ConfigMap
				for _, versionApp := range app.Spec.Versions {
					for _, module := range versionApp.Modules {
						if module.ConfigMapRef.Name != "" && module.ConfigMapRef.Namespace != "" {
							if _, err := c.helper.V1.ConfigMap.Delete(context.TODO(), module.ConfigMapRef.Namespace, module.ConfigMapRef.Name); err != nil {
								log.Error(err)
								return
							}
						}
						if module.AdditionalConfigMapRef.Name != "" && module.AdditionalConfigMapRef.Namespace != "" {
							if _, err := c.helper.V1.ConfigMap.Delete(context.TODO(), module.AdditionalConfigMapRef.Namespace, module.AdditionalConfigMapRef.Name); err != nil {
								log.Error(err)
								return
							}
						}
					}
				}
			case core.FinalizerCleanRefEvent:
				// 同步删除关联的事件
				eventList, err := c.helper.V1.Event.List(context.TODO(), "")
				if err != nil {
					log.Error(err)
					return
				}
				for _, eventObj := range eventList {
					event := eventObj.(*v1.Event)
					if event.Spec.ResourceRef.Kind == core.KindApp && event.Spec.ResourceRef.Name == app.Metadata.Name {
						if _, err := c.helper.V1.Event.Delete(context.TODO(), "", event.Metadata.Name, core.WithSync()); err != nil {
							log.Error(err)
							return
						}
					}
				}
			}

			c.deletings.Unset(app.GetKey())
			app.Metadata.Finalizers = app.Metadata.Finalizers[1:]
			if _, err := c.helper.V1.App.Update(context.TODO(), app, core.WithFinalizer()); err != nil {
				log.Error(err)
				return
			}
		} else {
			if _, err := c.helper.V1.App.Delete(context.TODO(), app.Metadata.Namespace, app.Metadata.Name); err != nil {
				log.Error(err)
				return
			}
		}
	}
}

func NewAppOperator() *AppOperator {
	o := &AppOperator{
		BaseOperator: NewBaseOperator(v1.NewAppRegistry()),
	}
	o.SetHandleFunc(o.handleApp)
	return o
}
