/*
 ** 此处存放v2版本资源结构的定义
 */

package v2

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
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

// SpecEncode 序列化Spec字段的内容
func (obj AppInstance) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

// SpecDecode 反序列化Spec字段的内容
func (obj *AppInstance) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

// SpecHash 计算Spec字段中的"有效"内容哈希值
func (obj AppInstance) SpecHash() string {
	for moduleIndex, module := range obj.Spec.Modules {
		for replicaIndex := range module.Replicas {
			obj.Spec.Modules[moduleIndex].Replicas[replicaIndex].Notes = ""
		}
	}
	obj.Spec.LivenessProbe = LivenessProbe{}
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// SpecEncode 序列化Spec字段的内容
func (obj Host) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

// SpecDecode 反序列化Spec字段的内容
func (obj *Host) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

// SpecHash 计算Spec字段中的"有效"内容哈希值
func (obj Host) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// SpecEncode 序列化Spec字段的内容
func (obj Job) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

// SpecDecode 反序列化Spec字段的内容
func (obj *Job) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

// SpecHash 计算Spec字段中的"有效"内容哈希值
func (obj Job) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// GetModule 根据名称获取模块
func (obj AppInstance) GetModule(moduleName string) (AppInstanceModule, bool) {
	for _, module := range obj.Spec.Modules {
		if module.Name == moduleName {
			return module, true
		}
	}
	return AppInstanceModule{}, false
}

// GetModule 根据参数名获取模块切片中的指定参数
func (obj AppInstance) GetModuleReplicaArgValue(moduleName string, replicaIndex int, argName string) interface{} {
	for _, module := range obj.Spec.Modules {
		if module.Name != moduleName {
			continue
		}
		for index, replica := range module.Replicas {
			if index != replicaIndex {
				continue
			}
			for _, arg := range replica.Args {
				if arg.Name == argName {
					return arg.Value
				}
			}
		}
	}
	return nil
}

// GetGlobalArgValue 根据参数名获取指定的全局参数
func (obj AppInstance) GetGlobalArgValue(argName string) interface{} {
	for _, arg := range obj.Spec.Global.Args {
		if arg.Name == argName {
			return arg.Value
		}
	}
	return nil
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

// NewHost 实例化主机
func NewHost() *Host {
	host := new(Host)
	host.Init(ApiVersion, core.KindHost)
	return host
}

// NewAppInstance 实例化应用实例
func NewAppInstance() *AppInstance {
	appInstance := new(AppInstance)
	appInstance.Init(ApiVersion, core.KindAppInstance)
	appInstance.Spec.LivenessProbe.InitialDelaySeconds = 10
	appInstance.Spec.LivenessProbe.PeriodSeconds = 60
	appInstance.Spec.LivenessProbe.TimeoutSeconds = 60
	return appInstance
}

// NewJob 实例化任务
func NewJob() *Job {
	job := new(Job)
	job.Init(ApiVersion, core.KindJob)
	job.Spec.TimeoutSeconds = core.JobDefaultTimeoutSeconds
	job.Spec.FailureThreshold = core.JobDefaultFailureThreshold
	return job
}
