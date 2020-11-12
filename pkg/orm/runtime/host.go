package runtime

import (
	"github.com/wujie1993/waves/pkg/orm/core"
)

type Host struct {
	core.BaseRuntimeObj `json:",inline" yaml:",inline"`
	Spec                HostSpec `json:"spec" yaml:"spec"`
}

type HostSpec struct {
	SSH HostSSH `json:"ssh" yaml:"ssh"`
}

type HostSSH struct {
	Host   string `json:"host" yaml:"host"`
	User   string `json:"user" yaml:"user"`
	Passwd string `json:"passwd" yaml:"passwd"`
}
