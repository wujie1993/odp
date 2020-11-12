package schedule

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/operators"
	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v2"
)

// Scheduler 任务调度器
type Scheduler struct {
	//workers     map[string]*Worker
	//cancels     map[string]context.CancelFunc
	workers     operators.MutexMap
	cancels     operators.MutexMap
	actionQueue chan core.ApiObjectAction
	ctx         context.Context
	helper      orm.Helper
	mutex       sync.Mutex
}

// Run 运行任务调度器
func (s *Scheduler) Run(ctx context.Context) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 把所有处于运行状态的任务置为Waiting状态
	objs, err := s.helper.V2.Job.List(context.TODO(), "")
	if err != nil {
		log.Error(err)
		return
	}
	for _, obj := range objs {
		job := obj.(*v2.Job)
		if job.Status.Phase == core.PhaseRunning {
			if _, err := s.helper.V2.Job.UpdateStatusPhase("", job.Metadata.Name, core.PhaseWaiting); err != nil {
				log.Error(err)
				return
			}
		}
	}

	s.ctx = ctx

	watchCtx, _ := context.WithCancel(s.ctx)
	watcher := s.helper.V2.Job.ListWatch(watchCtx, "")

	for {
		select {
		case <-s.ctx.Done():
			return
		case jobAction := <-watcher:
			log.Tracef("receive action '%+v' with content: %+v", jobAction.Type, jobAction.Obj)
			switch jobAction.Type {
			case db.KVActionTypeSet:
				s.handleAction(jobAction)
			case db.KVActionTypeDelete:
				key := jobAction.Obj.GetKey()
				if _, ok := s.workers.Get(key); ok {
					cancel, _ := s.cancels.Get(key)
					cancel.(context.CancelFunc)()
					log.Warnf("job canceled: %s", key)

					s.workers.Unset(key)
					s.cancels.Unset(key)
				}
			}
		case jobAction := <-s.actionQueue:
			s.handleAction(jobAction)
		}
	}
}

func (s *Scheduler) handleAction(jobAction core.ApiObjectAction) {
	key := jobAction.Obj.GetKey()
	job := jobAction.Obj.(*v2.Job)
	if job.Status.Phase != core.PhaseWaiting {
		return
	}
	if obj, ok := s.workers.Get(key); ok {
		worker := obj.(*Worker)
		if worker.IsBusy() {
			s.actionQueue <- jobAction
		} else if worker.Hash != job.SpecHash() {
			s.handleJob(job, worker)
		}
	} else {
		worker := new(Worker)
		s.handleJob(job, worker)
	}
}

func (s *Scheduler) handleJob(job *v2.Job, worker *Worker) {
	key := job.GetKey()
	worker.Hash = job.SpecHash()
	go func() {
		if _, err := s.helper.V2.Job.UpdateStatusPhase(job.Metadata.Namespace, job.Metadata.Name, core.PhaseRunning); err != nil {
			log.Println(err)
		}
		ctx, cancel := context.WithTimeout(s.ctx, job.Spec.TimeoutSeconds*time.Second)
		s.workers.Set(key, worker)
		s.cancels.Set(key, cancel)

		defer func() {
			s.workers.Unset(key)
			s.cancels.Unset(key)
		}()

		var err error
		for retry := 0; retry < job.Spec.FailureThreshold; retry++ {
			if err = worker.Run(ctx, job); err != nil {
				log.Error(err)
				continue
			}
			break
		}
		if err != nil {
			job.Status.SetCondition(core.ConditionTypeRun, err.Error())
			job.SetStatusPhase(core.PhaseFailed)
			if _, err := s.helper.V2.Job.Update(context.TODO(), job, core.WithStatus()); err != nil {
				log.Println(err)
			}
			return
		}

		job.Status.SetCondition(core.ConditionTypeInitialized, core.ConditionStatusTrue)
		job.SetStatusPhase(core.PhaseCompleted)
		if _, err := s.helper.V2.Job.UpdateStatus(job.Metadata.Namespace, job.Metadata.Name, job.Status); err != nil {
			log.Println(err)
			return
		}
	}()
}

func NewScheduler() *Scheduler {
	s := new(Scheduler)
	s.helper = orm.GetHelper()
	s.workers = operators.NewMutexMap()
	s.cancels = operators.NewMutexMap()
	s.actionQueue = make(chan core.ApiObjectAction, 1000)
	return s
}
