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
	Hash      string
	Revision  int
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
	HostAliases  []AppInstanceHostAliases
	ConfigMapRef ConfigMapRef
}

type AppInstanceModule struct {
	Name       string
	AppVersion string
	Replicas   []AppInstanceModuleReplica
}

type AppInstanceModuleReplica struct {
	Args              []AppInstanceArgs
	HostRefs          []string
	HostAliases       []AppInstanceHostAliases
	Notes             string
	ConfigMapRef      ConfigMapRef
	AdditionalConfigs AdditionalConfigs
}

type AdditionalConfigs struct {
	Enabled      bool
	ConfigMapRef ConfigMapRef
	Args         []AppInstanceArgs
}

type AppInstanceHostAliases struct {
	Hostname string
	IP       string
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

type Host struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            HostSpec
	Info            HostInfo
}

type HostSpec struct {
	SSH HostSSH
}

type HostSSH struct {
	Host     string
	User     string
	Password string
	Port     uint16
}

type HostInfo struct {
	OS      OS
	CPU     CPU
	Memory  Memory
	Disk    Disk
	GPUs    []GPUInfo
	Plugins []HostPlugin
}

type OS struct {
	Release string
	Kernel  string
}

type CPU struct {
	Cores int
	Model string
}

type Memory struct {
	Size  int
	Model string
}

type Disk struct {
	Size int
}

type AppInstanceRef struct {
	Namespace string
	Name      string
}

type GPUInfo struct {
	ID     int
	Model  string
	UUID   string
	Memory int
	Type   string
}

type HostPlugin struct {
	AppInstanceRef AppInstanceRef
	AppRef         AppRef
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
