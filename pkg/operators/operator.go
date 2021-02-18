package operators

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

const (
	AnnotationShortName = "ShortName"
)

// Event 管理器中记录事件日志时使用的结构
type Event struct {
	core.BaseApiObj
	Action string
	Msg    string
	JobRef string
	Phase  string
}

// MutexMap 是通过锁机制实现的协程安全字典
type MutexMap struct {
	hashMap map[string]interface{}
	mutex   sync.RWMutex
}

// Set 向字典中添加记录，当记录已存在时会发生覆盖
func (m *MutexMap) Set(key string, value interface{}) {
	m.mutex.Lock()
	m.hashMap[key] = value
	m.mutex.Unlock()
}

// Unset 从字典中移除记录
func (m *MutexMap) Unset(key string) {
	m.mutex.Lock()
	delete(m.hashMap, key)
	m.mutex.Unlock()
}

// Get 获取字典中的记录
func (m *MutexMap) Get(key string) (interface{}, bool) {
	m.mutex.RLock()
	value, ok := m.hashMap[key]
	m.mutex.RUnlock()
	return value, ok
}

// NewMutexMap 创建一个新的协程安全字典
func NewMutexMap() *MutexMap {
	return &MutexMap{
		hashMap: make(map[string]interface{}),
	}
}

// HandleFunc 资源处理方法定义
type HandleFunc func(ctx context.Context, obj core.ApiObject) error

// ReconcileFunc 收敛方法定义
type ReconcileFunc func(ctx context.Context, obj core.ApiObject)

// BaseOperator 基础管理器中实现了管理的生命周期管理，其中封装了各个资源管理器的通用字段与方法，可根据需要注入自定义的handler,reconciler和finalizer方法。
type BaseOperator struct {
	helper                *orm.Helper
	registry              registry.ApiObjectRegistry
	handle                HandleFunc
	reconcile             ReconcileFunc
	finalize              HandleFunc
	reconcilePeriodSecond int
	objQueue              chan core.ApiObject
	runMutex              sync.Mutex
	deletings             *MutexMap
	applyings             *MutexMap
}

// Run 运行资源管理器
func (o BaseOperator) Run(ctx context.Context) {
	// 开启分布式锁
	lockKey := o.getLockKey()
	if err := db.KV.Lock(context.TODO(), lockKey); err != nil {
		log.Error(err)
		return
	}
	defer db.KV.Unlock(context.TODO(), lockKey)

	// 运行并等待Reconcile与Handle退出
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go o.runReconcile(ctx, wg)
	go o.runHandle(ctx, wg)
	log.Debugf("%s operator is running", o.registry.GVK().Kind)
	wg.Wait()
}

// runReconcile 定时执行自定义的reconcile收敛逻辑，使资源达到理想中的状态
func (o BaseOperator) runReconcile(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	if o.reconcile == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			log.Debugf("%+v reconcile stopped", o.registry.GVK())
			return
		default:
			objs, err := o.registry.List(context.TODO(), "")
			if err != nil {
				log.Error(err)
			}
			for _, obj := range objs {
				reconcileCtx, _ := context.WithCancel(ctx)
				o.reconcile(reconcileCtx, obj)
			}
		}
		time.Sleep(time.Duration(o.reconcilePeriodSecond) * time.Second)
	}
}

// runHandle 开启资源变更的监听，当资源发生创建，更新或删除时会触发自定义的handle处理逻辑
func (o BaseOperator) runHandle(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	if o.handle == nil {
		return
	}

	watchCtx, _ := context.WithCancel(ctx)
	watcher := o.registry.ListWatch(watchCtx, "")

	for {
		select {
		case <-ctx.Done():
			log.Debugf("%+v handle stopped", o.registry.GVK())
			return
		case objAction, ok := <-watcher:
			if !ok {
				log.Warnf("%+v action watcher closed", o.registry.GVK())
				return
			}
			if objAction.Type == db.KVActionTypeSet && objAction.Obj != nil {
				handleCtx, _ := context.WithCancel(ctx)
				go func() {
					defer func() {
						e := recover()
						if err, ok := e.(error); ok {
							log.Error(err)
						} else if e != nil {
							log.Error(e)
						}
					}()
					o.handle(handleCtx, objAction.Obj)
				}()
			}
		case obj, ok := <-o.objQueue:
			if !ok {
				log.Errorf("%+v action queue closed", o.registry.GVK())
				return
			}
			if obj != nil {
				handleCtx, _ := context.WithCancel(ctx)
				go func() {
					defer func() {
						e := recover()
						if err, ok := e.(error); ok {
							log.Error(err)
						} else {
							log.Error(e)
						}
					}()
					o.handle(handleCtx, obj)
				}()
			}
		}
	}
}

// SetHandleFunc 设置自定义资源变更处理防范
func (o *BaseOperator) SetHandleFunc(f HandleFunc) {
	o.handle = f
}

// SetReconcileFunc 设置自定义定时收敛方法
func (o *BaseOperator) SetReconcileFunc(f ReconcileFunc, periodSecond int) {
	o.reconcile = f
	o.reconcilePeriodSecond = periodSecond
}

// SetReconcileFunc 设置自定义级联资源清理方法
func (o *BaseOperator) SetFinalizeFunc(f HandleFunc) {
	o.finalize = f
}

// delete 删除资源以及清理其关联资源
// 首先调用finalizer方法清理关联资源，关联资源的清理进度通过资源的Metadata.Finalizers字段表示
// 每次清理时从Metadata.Finalizers中取出第一项并清理其对应的资源，直至为空时，才实际删除资源记录。
func (o BaseOperator) delete(ctx context.Context, obj core.ApiObject) error {
	key := obj.GetKey()

	// 当资源未开始删除时初始化删除锁
	var mutex *sync.Mutex
	mutexObj, ok := o.deletings.Get(key)
	if !ok {
		mutex = &sync.Mutex{}
		o.deletings.Set(key, mutex)
	} else {
		mutex = mutexObj.(*sync.Mutex)
	}
	// 在资源开始一次新的finalize清理动作时，锁定资源，在清理结束后才释放接收下一次的finalize
	mutex.Lock()
	defer mutex.Unlock()

	metadata := obj.GetMetadata()

	if len(metadata.Finalizers) > 0 && o.finalize != nil {
		log.Tracef("finalizing %s of %s", metadata.Finalizers[0], obj.GetKey())

		// 每次只处理一项Finalizer
		if err := o.finalize(ctx, obj); err != nil {
			log.Error(err)
			return err
		}

		metadata.Finalizers = metadata.Finalizers[1:]
		obj.SetMetadata(metadata)
		if _, err := o.registry.Update(context.TODO(), obj, core.WithFinalizer()); err != nil {
			log.Error(err)
			return err
		}
	} else if len(metadata.Finalizers) == 0 {
		if _, err := o.registry.Delete(context.TODO(), metadata.Namespace, metadata.Name); err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}

// getLockKey 获取分布式锁键名
func (o BaseOperator) getLockKey() string {
	return core.RegistryPrefix + "/locks/" + o.registry.GVK().Kind
}

// recordEvent 记录事件日志
func (o BaseOperator) recordEvent(event Event) error {
	e := v1.NewEvent()
	e.Spec.ResourceRef.Kind = event.Kind
	e.Spec.ResourceRef.Namespace = event.Metadata.Namespace
	e.Spec.ResourceRef.Name = event.Metadata.Name
	e.Metadata.Annotations["ShortName"] = event.Metadata.Annotations["ShortName"]
	e.Spec.Action = event.Action
	e.Spec.Msg = event.Msg
	e.Spec.JobRef = event.JobRef
	e.Status.Phase = event.Phase
	return o.helper.V1.Event.Record(e)
}

// NewBaseOperator 创建基础管理器
func NewBaseOperator(r registry.ApiObjectRegistry) BaseOperator {
	return BaseOperator{
		registry:  r,
		helper:    orm.GetHelper(),
		objQueue:  make(chan core.ApiObject, 1000),
		applyings: NewMutexMap(),
		deletings: NewMutexMap(),
	}
}
