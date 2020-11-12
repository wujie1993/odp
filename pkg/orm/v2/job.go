package v2

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
	"github.com/wujie1993/waves/pkg/util"
)

func (obj Job) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

func (obj *Job) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

func (obj Job) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

type JobRegistry struct {
	registry.Registry
}

func (r JobRegistry) GetLogPath(jobsDir string, jobName string) (string, error) {
	obj, err := r.Get(context.TODO(), "", jobName)
	if err != nil {
		return "", err
	}
	if obj == nil {
		return "", nil
	}
	job := obj.(*Job)
	return path.Join(jobsDir, job.Metadata.Uid, "ansible.log"), nil
}

func (r JobRegistry) GetLog(jobsDir string, jobName string) ([]byte, error) {
	jobPath, err := r.GetLogPath(jobsDir, jobName)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(jobPath)
}

func (r JobRegistry) WatchLog(ctx context.Context, jobsDir string, jobName string) (<-chan string, error) {
	jobPath, err := r.GetLogPath(jobsDir, jobName)
	if err != nil {
		return nil, err
	}
	return util.Tailf(ctx, jobPath)
}

func jobPreCreate(obj core.ApiObject) error {
	job := obj.(*Job)
	job.Metadata.Finalizers = []string{core.FinalizerCleanRefConfigMap, core.FinalizerCleanJobWorkDir}
	return nil
}

func NewJob() *Job {
	job := new(Job)
	job.Init(ApiVersion, core.KindJob)
	job.Spec.TimeoutSeconds = core.JobDefaultTimeoutSeconds
	job.Spec.FailureThreshold = core.JobDefaultFailureThreshold
	return job
}

func NewJobRegistry() JobRegistry {
	r := JobRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindJob), false),
	}
	r.SetPreCreateHook(jobPreCreate)
	return r
}
