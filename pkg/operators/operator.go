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

type Event struct {
	core.BaseApiObj
	Action string
	Msg    string
	JobRef string
	Phase  string
}

type MutexMap struct {
	hashMap map[string]interface{}
	mutex   sync.RWMutex
}

func (m *MutexMap) Set(key string, value interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.hashMap[key] = value
}

func (m *MutexMap) Unset(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.hashMap, key)
}

func (m *MutexMap) Get(key string) (interface{}, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	value, ok := m.hashMap[key]
	return value, ok
}

func NewMutexMap() MutexMap {
	return MutexMap{
		hashMap: make(map[string]interface{}),
	}
}

type HandleFunc func(ctx context.Context, obj core.ApiObject)

type ReconcileFunc func(ctx context.Context, obj core.ApiObject)

type BaseOperator struct {
	helper                orm.Helper
	registry              registry.ApiObjectRegistry
	namespace             string
	handle                HandleFunc
	reconcile             ReconcileFunc
	reconcilePeriodSecond int
	objQueue              chan core.ApiObject
	runMutex              sync.Mutex
	deletings             MutexMap
	applyings             MutexMap
}

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

func (o BaseOperator) runReconcile(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	if o.reconcile == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			log.Debugf("%+v reconcile stopped", o.registry.GVK)
			return
		default:
			objs, err := o.registry.List(context.TODO(), o.namespace)
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

func (o BaseOperator) runHandle(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	if o.handle == nil {
		return
	}

	watchCtx, _ := context.WithCancel(ctx)
	watcher := o.registry.ListWatch(watchCtx, o.namespace)

	for {
		select {
		case <-ctx.Done():
			log.Debugf("%+v handle stopped", o.registry.GVK)
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

func (o *BaseOperator) SetHandleFunc(f HandleFunc) {
	o.handle = f
}

func (o *BaseOperator) SetReconcileFunc(f ReconcileFunc, periodSecond int) {
	o.reconcile = f
	o.reconcilePeriodSecond = periodSecond
}

func (o BaseOperator) getLockKey() string {
	return core.RegistryPrefix + "/locks/" + o.registry.GVK().Kind
}

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

func NewBaseOperator(r registry.ApiObjectRegistry) BaseOperator {
	return BaseOperator{
		namespace: core.DefaultNamespace,
		registry:  r,
		helper:    orm.GetHelper(),
		objQueue:  make(chan core.ApiObject, 1000),
		applyings: NewMutexMap(),
		deletings: NewMutexMap(),
	}
}
