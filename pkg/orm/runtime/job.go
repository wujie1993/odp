package runtime

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/wujie1993/waves/pkg/orm/core"
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

func NewJob() *Job {
	job := new(Job)
	job.Init("", core.KindJob)
	job.Spec.TimeoutSeconds = core.JobDefaultTimeoutSeconds
	job.Spec.FailureThreshold = core.JobDefaultFailureThreshold
	return job
}
