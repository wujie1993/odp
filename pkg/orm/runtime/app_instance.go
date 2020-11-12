package runtime

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/wujie1993/waves/pkg/orm/core"
)

func (obj AppInstance) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

func (obj *AppInstance) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

func (obj AppInstance) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func NewAppInstance() *AppInstance {
	appInstance := new(AppInstance)
	appInstance.Init("", core.KindAppInstance)
	appInstance.Spec.LivenessProbe.InitialDelaySeconds = 10
	appInstance.Spec.LivenessProbe.PeriodSeconds = 30
	appInstance.Spec.LivenessProbe.TimeoutSeconds = 30
	return appInstance
}
