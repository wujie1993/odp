package v1

import (
	"github.com/wujie1993/waves/pkg/orm/core"
)

const (
	ApiVersion = "v1"
)

type ValueFrom struct {
	ConfigMapRef ConfigMapRef
	HostRef      string
}

type ResourceRef struct {
	Namespace string
	Name      string
	Kind      string
}

type AppRef struct {
	Name    string
	Version string
}

type AppInstanceRef struct {
	Namespace string
	Name      string
}

type ConfigMapRef struct {
	Namespace string
	Name      string
}

type LivenessProbe struct {
	InitialDelaySeconds int
	PeriodSeconds       int
	TimeoutSeconds      int
}

type GPUInfo struct {
	ID     int
	Model  string
	UUID   string
	Memory int
	Type   string
}

func newGVK(kind string) core.GVK {
	return core.GVK{
		Group:      core.Group,
		ApiVersion: ApiVersion,
		Kind:       kind,
	}
}
