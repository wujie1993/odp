package v2

import (
	"time"

	"github.com/wujie1993/waves/pkg/orm/core"
)

const (
	ApiVersion = "v2"
)

type AppRef struct {
	Name    string
	Version string
}

type ConfigMapRef struct {
	Namespace string
	Name      string
	Key       string
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
	Args                   []AppInstanceArgs
	HostRefs               []string
	Notes                  string
	ConfigMapRef           ConfigMapRef
	AdditionalConfigMapRef ConfigMapRef
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

func (s AppInstanceSpec) GetModuleReplicaArgValue(moduleName string, replicaIndex int, argName string) (interface{}, bool) {
	for _, module := range s.Modules {
		if module.Name == moduleName {
			if replicaIndex >= len(module.Replicas) || replicaIndex < 0 {
				return nil, false
			}
			for _, arg := range module.Replicas[replicaIndex].Args {
				if arg.Name == argName {
					return arg.Value, true
				}
			}
		}
	}
	return nil, false
}

func (s *AppInstanceSpec) SetModuleReplicaArgValue(moduleName string, replicaIndex int, argName string, argValue interface{}) bool {
	for moduleIndex, module := range s.Modules {
		if module.Name == moduleName {
			if replicaIndex >= len(module.Replicas) || replicaIndex < 0 {
				return false
			}
			for argIndex, arg := range module.Replicas[replicaIndex].Args {
				if arg.Name == argName {
					s.Modules[moduleIndex].Replicas[replicaIndex].Args[argIndex].Value = argValue
					return true
				}
			}
		}
	}
	return false
}

func (s AppInstanceSpec) GetGlobalArgValue(argName string) (interface{}, bool) {
	for _, arg := range s.Global.Args {
		if arg.Name == argName {
			return arg.Value, true
		}
	}
	return nil, false
}

func (s *AppInstanceSpec) SetGlobalArgValue(argName string, argValue interface{}) bool {
	for argIndex, arg := range s.Global.Args {
		if arg.Name == argName {
			s.Global.Args[argIndex].Value = argValue
			return true
		}
	}
	return false
}
