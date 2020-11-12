package runtime

import (
	"time"

	"github.com/wujie1993/waves/pkg/orm/core"
)

type AppRef struct {
	Name    string
	Version string
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

type AppInstance struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            AppInstanceSpec
}

type AppInstanceSpec struct {
	Category      string
	AppRef        AppRef
	Action        string
	LivenessProbe LivenessProbe
	Modules       []AppInstanceModule
	Global        AppInstanceGlobal
	K8sRef        string
}

type AppInstanceGlobal struct {
	Args         []AppInstanceArgs
	ConfigMapRef ConfigMapRef
}

type AppInstanceModule struct {
	Name       string
	AppVersion string
	Replicas   []AppInstanceModuleReplica
}

type AppInstanceModuleReplica struct {
	Args         []AppInstanceArgs
	HostRefs     []string
	Notes        string
	ConfigMapRef ConfigMapRef
}

type AppInstanceArgs struct {
	Name  string
	Value interface{}
}

type ValueFrom struct {
	ConfigMapRef ConfigMapRef
}

type Job struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            JobSpec
}

type JobSpec struct {
	Exec             JobExec
	TimeoutSeconds   time.Duration
	FailureThreshold int
}

type JobExec struct {
	Type    string
	Ansible JobAnsible
}

type JobAnsible struct {
	Bin          string
	Plays        []JobAnsiblePlay
	RecklessMode bool
}

type JobAnsiblePlay struct {
	Name      string
	Envs      []string
	Tags      []string
	Playbook  AnsiblePlaybook
	Configs   []AnsibleConfig
	GroupVars AnsibleGroupVars
	Inventory AnsibleInventory
}

type AnsiblePlaybook struct {
	Value     string
	ValueFrom ValueFrom
}

type AnsibleGroupVars struct {
	Value     string
	ValueFrom ValueFrom
}

type AnsibleConfig struct {
	PathPrefix string
	ValueFrom  ValueFrom
}

type AnsibleInventory struct {
	Value     string
	ValueFrom ValueFrom
}
