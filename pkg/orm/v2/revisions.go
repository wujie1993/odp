package v2

import (
	"context"
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

// AppInstanceRevision 应用实例修订版本记录器，实现了Revisioner接口
type AppInstanceRevision struct {
	kind string
}

func (r AppInstanceRevision) SetRevision(ctx context.Context, obj core.ApiObject) error {
	appInstance := obj.(*AppInstance)

	// 如果与上个版本无差异，则不再创建新的历史版本
	lastRevision, err := r.GetLastRevision(ctx, appInstance.Metadata.Namespace, appInstance.Metadata.Name)
	if err != nil {
		return err
	}
	if lastRevision != nil && lastRevision.SpecHash() == obj.SpecHash() {
		return nil
	}

	// 只为原本状态为Installed的应用实例生成历史修订版本
	if appInstance.Status.Phase != core.PhaseInstalled {
		return nil
	}
	// 模块中的配置文件未标记版本时，不创建修订版本
	for _, module := range appInstance.Spec.Modules {
		for _, replica := range module.Replicas {
			if replica.ConfigMapRef.Name != "" && replica.ConfigMapRef.Hash == "" {
				return nil
			}
		}
	}

	revision := v1.NewRevision()
	revision.Metadata.Name = fmt.Sprintf("%s-%d-%s", appInstance.Metadata.Name, appInstance.Metadata.ResourceVersion, appInstance.SpecHash())
	revision.ResourceRef = v1.ResourceRef{
		Kind:      core.KindAppInstance,
		Namespace: appInstance.Metadata.Namespace,
		Name:      appInstance.Metadata.Name,
	}
	revision.Revision = appInstance.Metadata.ResourceVersion
	data, err := appInstance.ToJSON()
	if err != nil {
		return err
	}
	revision.Data = string(data)

	revisionRegistry := v1.NewRevisionRegistry()
	if _, err := revisionRegistry.Create(context.TODO(), revision); err != nil {
		return err
	}

	return nil
}

func (r AppInstanceRevision) ListRevisions(ctx context.Context, namespace string, name string) (core.ApiObjectList, error) {
	appInstanceRegistry := NewAppInstanceRegistry()

	obj, err := appInstanceRegistry.Get(ctx, namespace, name)
	if err != nil {
		return nil, err
	} else if obj == nil {
		return nil, nil
	}

	revisionRegistry := v1.NewRevisionRegistry()

	revisionList, err := revisionRegistry.List(context.TODO(), "")
	if err != nil {
		return nil, err
	}

	result := []core.ApiObject{}
	for _, revisionObj := range revisionList {
		revision := revisionObj.(*v1.Revision)
		if revision.ResourceRef.Kind == r.kind && revision.ResourceRef.Namespace == namespace && revision.ResourceRef.Name == name {
			item, err := New(r.kind)
			if err != nil {
				return nil, err
			}
			if err := item.FromJSON([]byte(revision.Data)); err != nil {
				return nil, err
			}

			result = append(result, item)
		}
	}

	sort.Sort(sort.Reverse(core.SortByRevision(result)))

	return result, nil
}

func (r AppInstanceRevision) GetRevision(ctx context.Context, namespace string, name string, revision int) (core.ApiObject, error) {
	appInstanceRegistry := NewAppInstanceRegistry()

	obj, err := appInstanceRegistry.Get(ctx, namespace, name)
	if err != nil {
		return nil, err
	} else if obj == nil {
		return nil, nil
	}
	appInstance := obj.(*AppInstance)
	if revision >= appInstance.Metadata.ResourceVersion {
		return nil, nil
	}

	revisionRegistry := v1.NewRevisionRegistry()

	revisionList, err := revisionRegistry.List(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, revisionObj := range revisionList {
		rev := revisionObj.(*v1.Revision)
		if rev.ResourceRef.Kind == r.kind && rev.ResourceRef.Namespace == namespace && rev.ResourceRef.Name == name && rev.Revision == revision {
			result, err := New(r.kind)
			if err != nil {
				return nil, err
			}
			if err := result.FromJSON([]byte(rev.Data)); err != nil {
				return nil, err
			}

			return result, nil
		}
	}
	return nil, nil
}

func (r AppInstanceRevision) RevertRevision(ctx context.Context, namespace string, name string, revision int) (core.ApiObject, error) {
	appInstanceRegistry := NewAppInstanceRegistry()

	obj, err := r.GetRevision(ctx, namespace, name, revision)
	if err != nil {
		log.Error(err)
		return nil, err
	} else if obj == nil {
		return nil, nil
	}

	appInstance := obj.(*AppInstance)
	appInstance.Spec.Action = core.AppActionRevert

	configMapRevision := v1.NewConfigMapRevision()
	// 回滚配置文件
	for _, module := range appInstance.Spec.Modules {
		for _, replica := range module.Replicas {
			if replica.ConfigMapRef.Name != "" {
				if _, err := configMapRevision.RevertRevision(ctx, replica.ConfigMapRef.Namespace, replica.ConfigMapRef.Name, replica.ConfigMapRef.Revision); err != nil {
					log.Error(err)
					return nil, err
				}
			}
		}
	}

	return appInstanceRegistry.Update(ctx, appInstance)
}

func (r AppInstanceRevision) GetLastRevision(ctx context.Context, namespace string, name string) (core.ApiObject, error) {
	objs, err := r.ListRevisions(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	if len(objs) > 0 {
		return objs[0], nil
	}
	return nil, nil
}

func (r AppInstanceRevision) DeleteRevision(ctx context.Context, namespace string, name string, revision int) (core.ApiObject, error) {
	appInstanceRegistry := NewAppInstanceRegistry()

	obj, err := appInstanceRegistry.Get(ctx, namespace, name)
	if err != nil {
		return nil, err
	} else if obj == nil {
		return nil, nil
	}

	resourceVersion := obj.GetMetadata().ResourceVersion
	if revision >= resourceVersion || resourceVersion <= 0 {
		return nil, nil
	}

	revisionRegistry := v1.NewRevisionRegistry()

	revisionList, err := revisionRegistry.List(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, revisionObj := range revisionList {
		rev := revisionObj.(*v1.Revision)
		if rev.ResourceRef.Kind == r.kind && rev.ResourceRef.Namespace == namespace && rev.ResourceRef.Name == name && rev.Revision == revision {
			result, err := New(r.kind)
			if err != nil {
				return nil, err
			}
			if err := result.FromJSON([]byte(rev.Data)); err != nil {
				return nil, err
			}

			if _, err := revisionRegistry.Delete(ctx, "", rev.Metadata.Name); err != nil {
				return nil, err
			}

			return result, nil
		}
	}
	return nil, nil
}

func (r AppInstanceRevision) DeleteAllRevisions(ctx context.Context, namespace string, name string) error {
	appInstanceRegistry := NewAppInstanceRegistry()

	obj, err := appInstanceRegistry.Get(ctx, namespace, name)
	if err != nil {
		return err
	} else if obj == nil {
		return nil
	}

	resourceVersion := obj.GetMetadata().ResourceVersion
	if resourceVersion <= 0 {
		return nil
	}

	revisionRegistry := v1.NewRevisionRegistry()

	revisionList, err := revisionRegistry.List(ctx, "")
	if err != nil {
		return err
	}

	for _, revisionObj := range revisionList {
		rev := revisionObj.(*v1.Revision)
		if rev.ResourceRef.Kind == r.kind && rev.ResourceRef.Namespace == namespace && rev.ResourceRef.Name == name {
			if _, err := revisionRegistry.Delete(ctx, "", rev.Metadata.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

// NewAppInstanceRevision 实例化应用实例修订版本记录器
func NewAppInstanceRevision() *AppInstanceRevision {
	return &AppInstanceRevision{
		kind: core.KindAppInstance,
	}
}
