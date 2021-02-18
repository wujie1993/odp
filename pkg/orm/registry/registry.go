package registry

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
)

// HookFunc 钩子方法定义
type HookFunc func(obj core.ApiObject) error

// ApiObjectRegistry 资源对象存储器接口，实现了该接口的对象可对资源对象进行数据库读写
type ApiObjectRegistry interface {
	// 写入一条新记录
	Create(ctx context.Context, obj core.ApiObject, opts ...core.OpOpt) (core.ApiObject, error)

	// 更新一条已存在的记录
	Update(ctx context.Context, obj core.ApiObject, opts ...core.OpOpt) (core.ApiObject, error)

	// 删除一条已存在的记录
	Delete(ctx context.Context, namespace string, name string, opts ...core.OpOpt) (core.ApiObject, error)

	// 获取一条已存在的记录
	Get(ctx context.Context, namespace, name string, opts ...core.OpOpt) (core.ApiObject, error)

	// 获取所有记录
	List(ctx context.Context, namespace string, opts ...core.OpOpt) (core.ApiObjectList, error)

	// 获取并监听所有记录的变更
	ListWatch(ctx context.Context, namespace string) <-chan core.ApiObjectAction

	// 将其他结构版本的记录转换成当前版本的结构并写入数据库
	MigrateObjects() error

	// 返回当前存储器对应的资源是否属于命名空间资源
	Namespaced() bool

	// 获取当前存储器所对应资源的分类，结构版本和类别
	GVK() core.GVK
}

// Registry 通用资源对象存储器，实现了通用的对象CRUD接口
type Registry struct {
	// 该存储器对应的资源对象类型
	gvk core.GVK

	// 该存储器对应的资源对象是否是命名空间资源
	namespaced bool

	// 创建与更新内容校验钩子
	validateHook HookFunc

	// 创建与更新内容填充钩子
	mutateHook HookFunc

	// 资源获取内容装饰钩子
	decorateHook HookFunc

	// 创建前置钩子
	preCreateHook HookFunc

	// 更新前置钩子
	preUpdateHook HookFunc

	// 删除前置钩子
	preDeleteHook HookFunc

	// 创建后置钩子
	postCreateHook HookFunc

	// 更新后置钩子
	postUpdateHook HookFunc

	// 删除后置钩子
	postDeleteHook HookFunc

	// 修订历史记录器
	revisioner Revisioner

	// 默认finalizers
	defaultFinalizers []string
}

// Create 创建单个资源对象
func (r Registry) Create(ctx context.Context, obj core.ApiObject, opts ...core.OpOpt) (core.ApiObject, error) {
	return r.createWithOpts(ctx, obj, opts...)
}

func (r Registry) createWithOpts(ctx context.Context, obj core.ApiObject, opts ...core.OpOpt) (core.ApiObject, error) {
	var option core.Option
	option.SetupOption(opts...)

	// 通用校验
	if err := r.commonValidate(obj); err != nil {
		return nil, err
	}

	// 执行自定义内容校验钩子
	if r.validateHook != nil {
		if err := r.validateHook(obj); err != nil {
			return nil, err
		}
	}

	// 执行自定义内容填充钩子
	if r.mutateHook != nil {
		if err := r.mutateHook(obj); err != nil {
			return nil, err
		}
	}

	metadata := obj.GetMetadata()

	// 获取存储键
	key := r.getKey(metadata.Namespace, metadata.Name)

	// 获取并判断对象是否存在
	if str, err := db.KV.Get(key); err != nil {
		return nil, err
	} else if str != "" {
		return nil, e.ResourceExistsError{Key: obj.GetKey()}
	}

	// 设置元数据
	if !r.namespaced {
		metadata.Namespace = ""
	}
	metadata.Uid = uuid.New().String()
	metadata.CreateTime = time.Now()
	metadata.UpdateTime = metadata.CreateTime
	metadata.Finalizers = r.defaultFinalizers
	obj.SetMetadata(metadata)
	obj.SetStatus(core.NewStatus())

	// 执行前置钩子
	if r.preCreateHook != nil {
		if err := r.preCreateHook(obj); err != nil {
			return nil, err
		}
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	// 创建对象
	if err := db.KV.Set(key, string(data)); err != nil {
		return nil, err
	}

	// 执行后置钩子
	if r.postCreateHook != nil {
		if err := r.postCreateHook(obj); err != nil {
			return nil, err
		}
	}
	log.Tracef("created %s: %s", obj.GetKey(), string(data))
	return obj, nil
}

// Update 更新单个资源对象
func (r Registry) Update(ctx context.Context, obj core.ApiObject, opts ...core.OpOpt) (core.ApiObject, error) {
	return r.updateWithOpts(ctx, obj, opts...)
}

func (r Registry) updateWithOpts(ctx context.Context, obj core.ApiObject, opts ...core.OpOpt) (core.ApiObject, error) {
	var option core.Option
	option.SetupOption(opts...)

	// 通用校验
	if err := r.commonValidate(obj); err != nil {
		return nil, err
	}

	// 执行自定义内容校验钩子
	if r.validateHook != nil {
		if err := r.validateHook(obj); err != nil {
			return nil, err
		}
	}

	// 执行自定义内容填充钩子
	if r.mutateHook != nil {
		if err := r.mutateHook(obj); err != nil {
			return nil, err
		}
	}

	// 执行更新前置钩子
	if r.preUpdateHook != nil {
		if err := r.preUpdateHook(obj); err != nil {
			return nil, err
		}
	}

	metadata := core.Metadata{}
	obj.GetMetadata().CopyTo(&metadata)

	key := r.getKey(metadata.Namespace, metadata.Name)

	// 获取并判断对象是否存在
	oldObj, err := r.Get(context.TODO(), metadata.Namespace, metadata.Name)
	if err != nil {
		return nil, err
	}
	if oldObj == nil {
		return nil, e.Errorf("update failed: %s not found", key)
	}

	// 更新或重置元数据
	oldMetadata := oldObj.GetMetadata()
	if !r.namespaced {
		metadata.Namespace = ""
	}
	metadata.CreateTime = oldMetadata.CreateTime
	metadata.UpdateTime = time.Now()
	metadata.Uid = oldMetadata.Uid
	metadata.ResourceVersion = oldMetadata.ResourceVersion

	// 启用WithAllFields选项时才可以更新Spec以外（不包括Metadata）的内容，否则会被重置
	if !option.WithAllFields {
		// 仅更新Spec
		spec, err := obj.SpecEncode()
		if err != nil {
			return nil, err
		}
		if err := core.DeepCopy(oldObj, obj); err != nil {
			return nil, err
		}
		if err := obj.SpecDecode(spec); err != nil {
			return nil, err
		}
	}

	oldSpec := oldObj.SpecHash()
	if obj.SpecHash() != oldSpec {
		// 资源内容体发生更新时累加资源版本号, 并将资源状态置为等待中
		metadata.Annotations[core.AnnotationLastAppliedConfiguration] = string(oldSpec)
		metadata.ResourceVersion++

		// 启用WithAllFields选项时才可以更新Status.Phase，否则会被重置
		if !option.WithAllFields {
			obj.SetStatusPhase(core.PhaseWaiting)
		}

		// 生成历史修订版本
		if r.revisioner != nil {
			if err := r.revisioner.SetRevision(ctx, oldObj); err != nil {
				return nil, err
			}
		}
	} else if option.WhenSpecChanged {
		return oldObj, nil
	}

	// 当没有指定WithFinalizer和WithAllFields中的其中一个选项时，重置Metadata.Finalizer
	if !option.WithFinalizer && !option.WithAllFields {
		// 当资源处于删除状态时，重置finalizers，否则还原至更新前的finalizers
		if obj.GetStatus().Phase != core.PhaseDeleting {
			metadata.Finalizers = r.defaultFinalizers
		} else {
			metadata.Finalizers = oldMetadata.Finalizers
		}
	}

	obj.SetMetadata(metadata)

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	// 更新对象
	if err := db.KV.Set(key, string(data)); err != nil {
		return nil, err
	}

	// 执行后置钩子
	if r.postUpdateHook != nil {
		if err := r.postUpdateHook(obj); err != nil {
			return nil, err
		}
	}
	log.Tracef("updated %s: %s", key, string(data))
	return obj, nil
}

// Delete 删除单个资源对象
func (r Registry) Delete(ctx context.Context, namespace string, name string, opts ...core.OpOpt) (core.ApiObject, error) {
	return r.deleteWithOpts(ctx, namespace, name, opts...)
}

func (r Registry) deleteWithOpts(ctx context.Context, namespace string, name string, opts ...core.OpOpt) (core.ApiObject, error) {
	var option core.Option
	option.SetupOption(opts...)

	// 字段校验
	re := regexp.MustCompile(core.ValidNameRegex)
	if r.namespaced && !re.MatchString(namespace) {
		err := e.InvalidNamespaceError{Namespace: namespace}
		log.Error(err)
		return nil, err
	}
	if !re.MatchString(name) {
		err := e.InvalidNameError{Name: name}
		log.Error(err)
		return nil, err
	}

	// 获取存储键
	key := r.getKey(namespace, name)

	// 获取并判断对象是否存在
	obj, err := r.Get(context.TODO(), namespace, name)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}

	// 执行前置钩子
	if r.preDeleteHook != nil {
		if err := r.preDeleteHook(obj); err != nil {
			return nil, err
		}
	}

	if len(obj.GetMetadata().Finalizers) > 0 {
		// 将对象置为删除中状态
		if _, err := r.UpdateStatusPhase(namespace, name, core.PhaseDeleting); err != nil {
			return nil, err
		}
	} else {
		// 删除对象
		if _, err := db.KV.Delete(key); err != nil {
			return nil, err
		}
	}

	// 执行后置钩子
	if r.postDeleteHook != nil {
		if err := r.postDeleteHook(obj); err != nil {
			return nil, err
		}
	}

	// 如果开启了同步删除，则等待直到资源删除完毕
	if option.WithSync {
		watcherCtx, _ := context.WithCancel(ctx)
		deleteWatcher := r.GetWatch(watcherCtx, namespace, name)
		if deleteWatcher != nil {
			for {
				select {
				case objAction, ok := <-deleteWatcher:
					if !ok {
						return obj, errors.New("delete channel close by unexcepted")
					}
					if objAction.Type == db.KVActionTypeDelete {
						return obj, nil
					}
				case <-ctx.Done():
					return obj, errors.New("delete channel close by unexcepted")
				}
			}
		}
	}

	return obj, nil
}

// Get 获取单个资源对象
func (r Registry) Get(ctx context.Context, namespace string, name string, opts ...core.OpOpt) (core.ApiObject, error) {
	return r.getWithOpts(ctx, namespace, name, opts...)
}

func (r Registry) getWithOpts(ctx context.Context, namespace string, name string, opts ...core.OpOpt) (core.ApiObject, error) {
	var option core.Option
	option.SetupOption(opts...)

	// 字段校验
	re := regexp.MustCompile(core.ValidNameRegex)
	if r.namespaced && !re.MatchString(namespace) {
		err := e.InvalidNamespaceError{Namespace: namespace}
		log.Error(err)
		return nil, err
	}
	if !re.MatchString(name) {
		err := e.InvalidNameError{Name: name}
		log.Error(err)
		return nil, err
	}

	// 获取存储键
	key := r.getKey(namespace, name)

	// 获取对象
	str, err := db.KV.Get(key)
	if err != nil {
		return nil, err
	}
	if str == "" {
		return nil, nil
	}

	// 判断对象版本
	metaType := new(core.MetaType)
	if err := json.Unmarshal([]byte(str), metaType); err != nil {
		return nil, err
	}
	var obj core.ApiObject
	// 解析对象
	if metaType.ApiVersion != r.gvk.ApiVersion {
		// 进行版本转换
		obj, err = convertByBytes([]byte(str), r.gvk)
		if err != nil {
			return nil, err
		}
	} else {
		// 直接解析对象
		obj, err = newByGVK(r.gvk)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(str), &obj); err != nil {
			return nil, err
		}
	}

	// 执行装饰钩子
	if r.decorateHook != nil {
		if err := r.decorateHook(obj); err != nil {
			return nil, err
		}
	}

	log.Tracef("got %s: %s", obj.GetKey(), str)
	return obj, nil
}

// List 列举单个命名空间下的所有资源对象
func (r Registry) List(ctx context.Context, namespace string, opts ...core.OpOpt) (core.ApiObjectList, error) {
	return r.listWithOpts(ctx, namespace, opts...)
}

func (r Registry) listWithOpts(ctx context.Context, namespace string, opts ...core.OpOpt) (core.ApiObjectList, error) {
	var option core.Option
	option.SetupOption(opts...)

	// 字段校验
	if namespace != "" {
		re := regexp.MustCompile(core.ValidNameRegex)
		if r.namespaced && !re.MatchString(namespace) {
			err := e.InvalidNamespaceError{Namespace: namespace}
			log.Error(err)
			return nil, err
		}
	}

	// 获取存储键
	key := r.getKey(namespace, "")

	// 获取对象
	kvList, err := db.KV.List(key, true)
	if err != nil {
		return nil, err
	}

	// 解析已获取的对象
	list := []core.ApiObject{}
	for _, value := range kvList {
		// 判断对象版本
		metaType := new(core.MetaType)
		if err := json.Unmarshal([]byte(value), metaType); err != nil {
			return nil, err
		}
		var obj core.ApiObject
		if metaType.ApiVersion != r.gvk.ApiVersion {
			// 存储版本与获取版本不一致，进行结构转换
			obj, err = convertByBytes([]byte(value), r.gvk)
			if err != nil {
				return nil, err
			}
		} else {
			obj, err = newByGVK(r.gvk)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal([]byte(value), &obj); err != nil {
				return nil, err
			}
		}

		// 执行装饰钩子
		if r.decorateHook != nil {
			if err := r.decorateHook(obj); err != nil {
				return nil, err
			}
		}

		list = append(list, obj)
	}
	log.Tracef("listed %s: %+v", key, list)
	return list, nil
}

// Watch 侦听资源对象的变动, 当name为空时，表示侦听命名空间下的所有资源对象
func (r Registry) Watch(ctx context.Context, namespace string, name string) <-chan core.ApiObjectAction {
	// 字段校验
	if namespace != "" {
		re := regexp.MustCompile(core.ValidNameRegex)
		if r.namespaced && !re.MatchString(namespace) {
			err := e.InvalidNamespaceError{Namespace: namespace}
			log.Error(err)
			return nil
		}
	}

	// 获取存储键
	key := r.getKey(namespace, name)

	// 根据是否传递名称判断是侦听单个还是所有对象
	var kvActionWatcher <-chan db.KVAction
	if name != "" {
		kvActionWatcher = db.KV.Watch(ctx, key, false)
	} else {
		kvActionWatcher = db.KV.Watch(ctx, key, true)
	}

	objActionChan := make(chan core.ApiObjectAction)
	go func() {
		defer close(objActionChan)
		for {
			select {
			case kvAction, ok := <-kvActionWatcher:
				if !ok {
					return
				}

				// 转换对象版本
				metaType := new(core.MetaType)
				if err := json.Unmarshal([]byte(kvAction.Value), metaType); err != nil {
					log.Error(err)
				}
				var obj core.ApiObject
				var err error
				if metaType.ApiVersion != r.gvk.ApiVersion {
					// 存储版本与获取版本不一致，进行结构转换
					obj, err = convertByBytes([]byte(kvAction.Value), r.gvk)
					if err != nil {
						log.Error(err)
					}
				} else {
					// 解析已侦听到的对象
					obj, err = newByGVK(r.gvk)
					if err != nil {
						log.Error(err)
					}
					if err := json.Unmarshal([]byte(kvAction.Value), &obj); err != nil {
						log.Error(err)
					}
				}

				// 执行装饰钩子
				if r.decorateHook != nil {
					if err := r.decorateHook(obj); err != nil {
						log.Error(err)
					}
				}

				// 将侦听到的对象推入响应通道
				objActionChan <- core.ApiObjectAction{
					Type: kvAction.ActionType,
					Obj:  obj,
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	log.Tracef("watched %s", key)
	return objActionChan
}

// GetWatch 侦听单个资源对象的变动
func (r Registry) GetWatch(ctx context.Context, namespace string, name string) <-chan core.ApiObjectAction {
	// 字段校验
	re := regexp.MustCompile(core.ValidNameRegex)
	if r.namespaced && !re.MatchString(namespace) {
		err := e.InvalidNamespaceError{Namespace: namespace}
		log.Error(err)
		return nil
	}
	if !re.MatchString(name) {
		err := e.InvalidNameError{Name: name}
		log.Error(err)
		return nil
	}

	// 获取存储键
	key := r.getKey(namespace, name)

	// 根据是否传递名称判断是侦听单个还是所有对象
	var kvActionWatcher <-chan db.KVAction
	kvActionWatcher = db.KV.Watch(ctx, key, false)

	objActionChan := make(chan core.ApiObjectAction, 1000)

	obj, _ := r.Get(context.TODO(), namespace, name)
	if obj != nil {
		objActionChan <- core.ApiObjectAction{
			Type: db.KVActionTypeSet,
			Obj:  obj,
		}
	} else {
		return nil
	}

	go func() {
		defer close(objActionChan)

		for {
			select {
			case kvAction, ok := <-kvActionWatcher:
				if !ok {
					return
				}

				// 转换对象版本
				metaType := new(core.MetaType)
				if err := json.Unmarshal([]byte(kvAction.Value), metaType); err != nil {
					log.Error(err)
				}
				var obj core.ApiObject
				var err error
				if metaType.ApiVersion != r.gvk.ApiVersion {
					// 存储版本与获取版本不一致，进行结构转换
					obj, err = convertByBytes([]byte(kvAction.Value), r.gvk)
					if err != nil {
						log.Error(err)
					}
				} else {
					// 解析已侦听到的对象
					obj, err = newByGVK(r.gvk)
					if err != nil {
						log.Error(err)
					}
					if err := json.Unmarshal([]byte(kvAction.Value), &obj); err != nil {
						log.Error(err)
					}
				}

				// 执行装饰钩子
				if r.decorateHook != nil {
					if err := r.decorateHook(obj); err != nil {
						log.Error(err)
					}
				}

				// 将侦听到的对象推入响应通道
				objActionChan <- core.ApiObjectAction{
					Type: kvAction.ActionType,
					Obj:  obj,
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	log.Tracef("watched %s", key)
	return objActionChan
}

// ListWatch 列举单个命名空间下的所有资源对象变动
func (r Registry) ListWatch(ctx context.Context, namespace string) <-chan core.ApiObjectAction {
	// 字段校验
	if namespace != "" {
		re := regexp.MustCompile(core.ValidNameRegex)
		if r.namespaced && !re.MatchString(namespace) {
			err := e.InvalidNamespaceError{Namespace: namespace}
			log.Error(err)
			return nil
		}
	}

	// 获取存储键
	key := r.getKey(namespace, "")

	// 根据是否传递名称判断是侦听单个还是所有对象
	var kvActionWatcher <-chan db.KVAction
	kvActionWatcher = db.KV.Watch(ctx, key, true)

	objActionChan := make(chan core.ApiObjectAction, 1000)

	go func() {
		defer close(objActionChan)

		list, _ := r.List(context.TODO(), namespace)
		for _, obj := range list {
			objActionChan <- core.ApiObjectAction{
				Type: db.KVActionTypeSet,
				Obj:  obj,
			}
		}

		for {
			select {
			case kvAction, ok := <-kvActionWatcher:
				if !ok {
					return
				}

				// 转换对象版本
				metaType := new(core.MetaType)
				if err := json.Unmarshal([]byte(kvAction.Value), metaType); err != nil {
					log.Error(err)
				}
				var obj core.ApiObject
				var err error
				if metaType.ApiVersion != r.gvk.ApiVersion {
					// 存储版本与获取版本不一致，进行结构转换
					obj, err = convertByBytes([]byte(kvAction.Value), r.gvk)
					if err != nil {
						log.Error(err)
					}
				} else {
					// 解析已侦听到的对象
					obj, err = newByGVK(r.gvk)
					if err != nil {
						log.Error(err)
					}
					if err := json.Unmarshal([]byte(kvAction.Value), &obj); err != nil {
						log.Error(err)
					}
				}

				// 执行装饰钩子
				if r.decorateHook != nil {
					if err := r.decorateHook(obj); err != nil {
						log.Error(err)
					}
				}

				// 将侦听到的对象推入响应通道
				objActionChan <- core.ApiObjectAction{
					Type: kvAction.ActionType,
					Obj:  obj,
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	log.Tracef("watched %s", key)
	return objActionChan
}

// UpdateStatus 更新单个资源对象的Status
func (r Registry) UpdateStatus(namespace string, name string, status core.Status) (core.ApiObject, error) {
	// 字段校验
	re := regexp.MustCompile(core.ValidNameRegex)
	if r.namespaced && !re.MatchString(namespace) {
		err := e.InvalidNamespaceError{Namespace: namespace}
		log.Error(err)
		return nil, err
	}
	if !re.MatchString(name) {
		err := e.InvalidNameError{Name: name}
		log.Error(err)
		return nil, err
	}

	// 获取存储键
	key := r.getKey(namespace, name)

	// 获取并判断对象是否存在
	obj, err := r.Get(context.TODO(), namespace, name)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, e.Errorf("update failed: %s not found", key)
	}

	obj.SetUpdateTime(time.Now())
	obj.SetStatus(status)

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	// 更新对象
	if err := db.KV.Set(key, string(data)); err != nil {
		return nil, err
	}
	log.Tracef("updated %s: %s", obj.GetKey(), string(data))
	return obj, nil
}

// UpdateStatusPhase 更新单个资源对象的Status.Phase
func (r Registry) UpdateStatusPhase(namespace string, name string, phase string) (core.ApiObject, error) {
	// 字段校验
	re := regexp.MustCompile(core.ValidNameRegex)
	if r.namespaced && !re.MatchString(namespace) {
		err := e.InvalidNamespaceError{Namespace: namespace}
		log.Error(err)
		return nil, err
	}
	if !re.MatchString(name) {
		err := e.InvalidNameError{Name: name}
		log.Error(err)
		return nil, err
	}

	// 获取存储键
	key := r.getKey(namespace, name)

	// 获取并判断对象是否存在
	obj, err := r.Get(context.TODO(), namespace, name)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, e.Errorf("update failed: %s not found", key)
	}

	obj.SetUpdateTime(time.Now())
	obj.SetStatusPhase(phase)

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	// 更新对象
	if err := db.KV.Set(key, string(data)); err != nil {
		return nil, err
	}
	log.Tracef("updated %s: %s", obj.GetKey(), string(data))
	return obj, nil
}

// MigrateObjects 将所有资源对象转换为当前存储器所指定的版本
func (r Registry) MigrateObjects() error {
	var key string

	// 获取存储键前缀
	if r.namespaced {
		key = r.getKey(core.DefaultNamespace, "")
	} else {
		key = r.getKey("", "")
	}

	// 获取对象
	kvList, err := db.KV.List(key, true)
	if err != nil {
		return err
	}

	// 解析已获取的对象
	for _, value := range kvList {
		// 判断对象版本
		metaType := new(core.MetaType)
		if err := json.Unmarshal([]byte(value), metaType); err != nil {
			return err
		}
		if metaType.Kind != r.gvk.Kind {
			continue
		}

		// 存储版本与获取版本不一致，进行结构转换
		obj, err := convertByBytes([]byte(value), r.gvk)
		if err != nil {
			return err
		}

		metadata := obj.GetMetadata()
		hash := obj.Sha256()

		// 修复资源内容
		if !r.namespaced {
			metadata.Namespace = ""
		}
		if obj.GetStatusPhase() != core.PhaseDeleting {
			metadata.Finalizers = r.defaultFinalizers
		}
		if r.mutateHook != nil {
			r.mutateHook(obj)
		}
		obj.SetMetadata(metadata)

		if metaType.ApiVersion != r.gvk.ApiVersion {
			// 结构版本不同时，存储转换后的结构内容
			log.Debugf("migrate %s from version %s to %s", obj.GetKey(), metaType.ApiVersion, r.gvk.ApiVersion)
		} else {
			// 对比资源修复前与修复后的内容是否发生变化，有变化则更新为修复后的内容

			// 结构版本相同时，如果无需填充数据，则跳过
			if hash == obj.Sha256() {
				continue
			}

			log.Debugf("update %s", obj.GetKey())
		}

		data, err := json.Marshal(obj)
		if err != nil {
			return err
		}

		if err := db.KV.Set(r.getKey(metadata.Namespace, metadata.Name), string(data)); err != nil {
			return err
		}
	}
	return nil
}

func (r Registry) GVK() core.GVK {
	return r.gvk
}

// SetValidateHook 注入用于自定义校验钩子，该钩子会在Create和Update前执行
func (r *Registry) SetValidateHook(hook HookFunc) {
	r.validateHook = hook
}

// SetMutateHook 注入自定义数据填充钩子，该钩子会在Create和Update前执行
func (r *Registry) SetMutateHook(hook HookFunc) {
	r.mutateHook = hook
}

// SetDecorateHook 注入获取对象后的内容填充钩子，该钩子会在Get和Watch获取到对象后执行
func (r *Registry) SetDecorateHook(hook HookFunc) {
	r.decorateHook = hook
}

// SetPreCreateHook 注入前置创建钩子
func (r *Registry) SetPreCreateHook(hook HookFunc) {
	r.preCreateHook = hook
}

// SetPreUpdateHook 注入前置更新钩子
func (r *Registry) SetPreUpdateHook(hook HookFunc) {
	r.preUpdateHook = hook
}

// SetPreCreateHook 注入前置删除钩子
func (r *Registry) SetPreDeleteHook(hook HookFunc) {
	r.preDeleteHook = hook
}

// SetPostCreateHook 注入后置创建钩子
func (r *Registry) SetPostCreateHook(hook HookFunc) {
	r.postCreateHook = hook
}

// SetPostUpdateHook 注入后置更新钩子
func (r *Registry) SetPostUpdateHook(hook HookFunc) {
	r.postUpdateHook = hook
}

// SetPostCreateHook 注入后置删除钩子
func (r *Registry) SetPostDeleteHook(hook HookFunc) {
	r.postDeleteHook = hook
}

// SetRevisioner 设置修订历史记录器
func (r *Registry) SetRevisioner(revisioner Revisioner) {
	r.revisioner = revisioner
}

// SetDefaultFinalizers 设置默认finalizers
func (r *Registry) SetDefaultFinalizers(finalizers []string) {
	r.defaultFinalizers = finalizers
}

// Namespaced 返回该存储器对应的是否是命名空间资源
func (r Registry) Namespaced() bool {
	return r.namespaced
}

func (r Registry) getKey(namespace string, name string) string {
	key := core.RegistryPrefix + "/" + r.gvk.Kind + "s/"
	if r.namespaced {
		if namespace != "" {
			key += namespace + "/"
			if name != "" {
				key += name
			}
		}
	} else {
		if name != "" {
			key += name
		}
	}
	return key
}

// 内容校验
func (r Registry) commonValidate(obj core.ApiObject) error {
	metadata := obj.GetMetadata()
	metatype := obj.GetMetaType()
	if r.gvk.ApiVersion != metatype.ApiVersion {
		return errors.New("apiVersion does not match with registry")
	}
	if r.gvk.Kind != metatype.Kind {
		return errors.New("kind does not match with registry")
	}
	re := regexp.MustCompile(core.ValidNameRegex)
	if r.namespaced && !re.MatchString(metadata.Namespace) {
		return e.InvalidNamespaceError{Namespace: metadata.Namespace}
	}
	if !re.MatchString(metadata.Name) {
		return e.InvalidNameError{Name: metadata.Name}
	}
	return nil
}

// NewRegistry 创建一个通用资源对象存储器
func NewRegistry(gvk core.GVK, namespaced bool) Registry {
	return Registry{
		gvk:        gvk,
		namespaced: namespaced,
	}
}
