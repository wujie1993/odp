package runtime

import (
	"github.com/wujie1993/waves/pkg/orm/core"
)

type ConfigMap struct {
	core.BaseRuntimeObj `json:",inline" yaml:",inline"`
	Spec                map[string]string `json:"spec" yaml:"spec"`
}
