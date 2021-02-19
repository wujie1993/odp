package v1

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wujie1993/waves/pkg/orm/core"
)

const (
	ApiVersion = "v1"
)

type App struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            AppSpec
}

type AppSpec struct {
	Category string
	Platform string
	Versions []AppVersion
}

type AppVersion struct {
	Version           string
	ShortName         string
	Platform          string
	Desc              string
	Increment         bool
	SupportActions    []string
	SupportMediaTypes []string
	SupportGpuModels  []string
	Enabled           bool
	DashboardId       string
	PkgRef            string
	LivenessProbe     LivenessProbe
	Modules           []AppModule
	Global            AppGlobal
	GpuRequired       bool
}

type AppModule struct {
	Name              string
	Desc              string
	SkipUpgrade       bool
	Required          bool
	Notes             string
	Replication       bool
	HostLimits        HostLimits
	IncludeRoles      []string
	Args              []AppArgs
	ConfigMapRef      ConfigMapRef
	AdditionalConfigs AdditionalConfigs
	EnableLogging     bool
	EnablePurgeData   bool
	ExtraVars         map[string]interface{}
	Resources         Resources
	HostAliases       []string
}

type AdditionalConfigs struct {
	Enabled      bool
	ConfigMapRef ConfigMapRef
	Args         []AppArgs
}

type AppGlobal struct {
	Args         []AppArgs
	ConfigMapRef ConfigMapRef
	HostAliases  []string
}

type HostLimits struct {
	Max int
	Min int
}

type AppArgs struct {
	Name       string
	ShortName  string
	Desc       string
	Type       string
	Format     string
	Enum       []string
	HostLimits HostLimits
	Default    interface{}
	Required   bool
	Modifiable bool
	Readonly   bool
}

type Resources struct {
	AlgorithmPlugin               bool
	SupportAlgorithmPluginsRegexp string
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
	Name              string
	Notes             string
	AppVersion        string
	Args              []AppInstanceArgs
	HostRefs          []string
	HostAliases       []AppInstanceHostAliases
	ConfigMapRef      ConfigMapRef
	AdditionalConfigs AdditionalConfigs
}

type AppInstanceHostAliases struct {
	Hostname string
	IP       string
}

type AppInstanceArgs struct {
	Name  string
	Value interface{}
}

type Audit struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            AuditSpec
}

type AuditSpec struct {
	ResourceRef ResourceRef
	Action      string
	Msg         string
	SourceIP    string
	ReqBody     string
	RespBody    string
	StatusCode  int
}

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
	Hash      string
	Revision  int
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

type ConfigMap struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Data            map[string]string
}

type Event struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            EventSpec
}

type EventSpec struct {
	ResourceRef ResourceRef
	Action      string
	Msg         string
	JobRef      string
}

type GPU struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            GPUSpec
}

type GPUSpec struct {
	HostRef              string
	Info                 GPUInfo
	AppInstanceModuleRef AppInstanceModuleRef
}

type Host struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            HostSpec
}

type HostSpec struct {
	SSH     HostSSH
	Info    HostInfo
	Plugins []HostPlugin
	Sdks    []SdkPlugin
}

type HostSSH struct {
	Host     string
	User     string
	Password string
	Port     uint16
}

type HostInfo struct {
	OS     OS
	CPU    CPU
	Memory Memory
	Disk   Disk
	GPUs   []GPUInfo
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

type HostPlugin struct {
	AppInstanceRef AppInstanceRef
	AppRef         AppRef
}

type SdkPlugin struct {
	AppInstanceRef AppInstanceRef
	AppRef         AppRef
}

type AppInstanceModuleRef struct {
	AppInstanceRef
	Module  string
	Replica int
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
	Inventories  []AnsibleInventory
	Envs         []string
	Tags         []string
	Playbook     string
	Configs      []JobConfig
	GroupVars    GroupVars
	RecklessMode bool
}

type GroupVars struct {
	ValueFrom ValueFrom
}

type JobConfig struct {
	Path         string
	ConfigMapRef ConfigMapRef
}

type AnsibleInventory struct {
	Value     string
	ValueFrom ValueFrom
}

type K8sConfig struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            K8sYaml
}

type K8sYaml struct {
	Action string
	Chrony string
	Etcd   struct {
		Hosts []K8sHostRef
	}
	ExLb   string
	Harbor struct {
		Hosts []K8sHostRef
	}
	K8SMaster struct {
		Hosts []K8sHostRef
	} `json:"K8s-master" yaml:"K8s-master"`
	K8SWorker struct {
		Hosts []K8sHostRef
	} `json:"K8s-worker" yaml:"K8s-worker"`
	K8SWorkerNew struct {
		Hosts []K8sHostRef
	} `json:"K8s-worker-new" yaml:"K8s-worker-new"`
	GPU struct {
		Hosts []K8sHostRef
	}
}

type K8sHostRef struct {
	ValueFrom ValueFrom
	Label     map[string]string
}

type Namespace struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
}

type Pkg struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            PkgSpec
}

type PkgSpec struct {
	Desc      string
	Module    string
	Version   string
	Platform  string
	Provision string
	Synced    bool
	Author    string
	Images    []string
}

type Project struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	ReferNamespaces []string
}

type Revision struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	ResourceRef     ResourceRef
	Revision        int
	Data            string
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
	for moduleIndex := range obj.Spec.Modules {
		obj.Spec.Modules[moduleIndex].Notes = ""
	}
	obj.Spec.LivenessProbe = LivenessProbe{}
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// SpecEncode 序列化Spec字段的内容
func (obj App) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

// SpecDecode 反序列化Spec字段的内容
func (obj *App) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

// SpecHash 计算Spec字段中的"有效"内容哈希值
func (obj App) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// SpecEncode 序列化Spec字段的内容
func (obj Audit) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

// SpecDecode 反序列化Spec字段的内容
func (obj *Audit) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

// SpecHash 计算Spec字段中的"有效"内容哈希值
func (obj Audit) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// SpecEncode 序列化Spec字段的内容
func (obj ConfigMap) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Data)
}

// SpecDecode 反序列化Spec字段的内容
func (obj *ConfigMap) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Data)
}

// SpecHash 计算Spec字段中的"有效"内容哈希值
func (obj ConfigMap) SpecHash() string {
	data, _ := json.Marshal(&obj.Data)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// SpecEncode 序列化Spec字段的内容
func (obj Event) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

// SpecDecode 反序列化Spec字段的内容
func (obj *Event) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

// SpecHash 计算Spec字段中的"有效"内容哈希值
func (obj Event) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// SpecEncode 序列化Spec字段的内容
func (obj GPU) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

// SpecDecode 反序列化Spec字段的内容
func (obj *GPU) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

// SpecHash 计算Spec字段中的"有效"内容哈希值
func (obj GPU) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// SpecEncode 序列化Spec字段的内容
func (obj Host) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec.SSH)
}

// SpecDecode 反序列化Spec字段的内容
func (obj *Host) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec.SSH)
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

// SpecEncode 序列化Spec字段的内容
func (obj K8sConfig) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

// SpecDecode 反序列化Spec字段的内容
func (obj *K8sConfig) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

// SpecHash 计算Spec字段中的"有效"内容哈希值
func (obj K8sConfig) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// SpecEncode 序列化Spec字段的内容
func (obj Namespace) SpecEncode() ([]byte, error) {
	return nil, nil
}

// SpecDecode 反序列化Spec字段的内容
func (obj *Namespace) SpecDecode(data []byte) error {
	return nil
}

// SpecHash 计算Spec字段中的"有效"内容哈希值
func (obj Namespace) SpecHash() string {
	return ""
}

// SpecEncode 序列化Spec字段的内容
func (obj Pkg) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

// SpecDecode 反序列化Spec字段的内容
func (obj *Pkg) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

// SpecHash 计算Spec字段中的"有效"内容哈希值
func (obj Pkg) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// SpecEncode 序列化Spec字段的内容
func (obj Project) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj)
}

// SpecDecode 反序列化Spec字段的内容
func (obj *Project) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj)
}

// SpecHash 计算Spec字段中的"有效"内容哈希值
func (obj Project) SpecHash() string {
	data, _ := json.Marshal(&obj)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// SpecEncode 序列化Spec字段的内容
func (obj Revision) SpecEncode() ([]byte, error) {
	return nil, nil
}

// SpecDecode 反序列化Spec字段的内容
func (obj *Revision) SpecDecode(data []byte) error {
	return nil
}

// SpecHash 计算Spec字段中的"有效"内容哈希值
func (obj Revision) SpecHash() string {
	return ""
}

func (s AppInstanceSpec) GetModuleArgValue(moduleName, argName string) (interface{}, bool) {
	for _, module := range s.Modules {
		if module.Name == moduleName {
			for _, arg := range module.Args {
				if arg.Name == argName {
					return arg.Value, true
				}
			}
		}
	}
	return nil, false
}

func (s *AppInstanceSpec) SetModuleArgValue(moduleName, argName string, argValue interface{}) bool {
	for moduleIndex, module := range s.Modules {
		if module.Name == moduleName {
			for argIndex, arg := range module.Args {
				if arg.Name == argName {
					s.Modules[moduleIndex].Args[argIndex].Value = argValue
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

func (obj App) VersionEnabled(version string) bool {
	for _, appVersion := range obj.Spec.Versions {
		if appVersion.Version == version {
			return appVersion.Enabled
		}
	}
	return false
}

func (obj App) GetVersion(version string) (AppVersion, bool) {
	for _, appVersion := range obj.Spec.Versions {
		if appVersion.Version == version {
			return appVersion, true
		}
	}
	return AppVersion{}, false
}

func (obj App) GetVersionModule(version string, moduleName string) (AppModule, bool) {
	appVersion, ok := obj.GetVersion(version)
	if !ok {
		return AppModule{}, false
	}
	return appVersion.GetModule(moduleName)
}

func (obj AppVersion) GetModule(moduleName string) (AppModule, bool) {
	for _, module := range obj.Modules {
		if module.Name == moduleName {
			return module, true
		}
	}
	return AppModule{}, false
}

// NewApp 实例化应用
func NewApp() *App {
	app := new(App)
	app.Init(ApiVersion, core.KindApp)
	app.Spec.Versions = []AppVersion{}
	return app
}

// NewAppInstance 实例化应用实例
func NewAppInstance() *AppInstance {
	appInstance := new(AppInstance)
	appInstance.Init(ApiVersion, core.KindAppInstance)
	appInstance.Spec.LivenessProbe.InitialDelaySeconds = 10
	appInstance.Spec.LivenessProbe.PeriodSeconds = 60
	appInstance.Spec.LivenessProbe.TimeoutSeconds = 60
	appInstance.Spec.Modules = []AppInstanceModule{}
	return appInstance
}

// NewAudit 实例化审计日志
func NewAudit() *Audit {
	audit := new(Audit)
	audit.Init(ApiVersion, core.KindAudit)
	return audit
}

// NewConfigMap 实例化配置字典
func NewConfigMap() *ConfigMap {
	configMap := new(ConfigMap)
	configMap.Init(ApiVersion, core.KindConfigMap)
	configMap.Data = make(map[string]string)
	return configMap
}

// NewEvent 实例化事件日志
func NewEvent() *Event {
	event := new(Event)
	event.Init(ApiVersion, core.KindEvent)
	return event
}

// NewGPU 实例化显卡
func NewGPU() *GPU {
	gpu := new(GPU)
	gpu.Init(ApiVersion, core.KindGPU)
	return gpu
}

// NewHost 实例化主机
func NewHost() *Host {
	host := new(Host)
	host.Init(ApiVersion, core.KindHost)
	host.Spec.Plugins = []HostPlugin{}
	host.Spec.Sdks = []SdkPlugin{}
	return host
}

// NewJob 实例化任务
func NewJob() *Job {
	job := new(Job)
	job.Init(ApiVersion, core.KindJob)
	job.Spec.TimeoutSeconds = core.JobDefaultTimeoutSeconds
	job.Spec.FailureThreshold = core.JobDefaultFailureThreshold
	return job
}

// NewK8sConfig 实例化K8S集群
func NewK8sConfig() *K8sConfig {
	k8sConfig := new(K8sConfig)
	k8sConfig.Init(ApiVersion, core.KindK8sConfig)
	return k8sConfig
}

// NewNamespace 实例化命名空间
func NewNamespace() *Namespace {
	ns := new(Namespace)
	ns.Init(ApiVersion, core.KindNamespace)
	return ns
}

// NewPkg 实例化部署包
func NewPkg() *Pkg {
	host := new(Pkg)
	host.Init(ApiVersion, core.KindPkg)
	return host
}

// NewProject 实例化项目
func NewProject() *Project {
	project := new(Project)
	project.Init(ApiVersion, core.KindProject)
	return project
}

// NewRevision 实例化修订版本
func NewRevision() *Revision {
	revision := new(Revision)
	revision.Init(ApiVersion, core.KindRevision)
	return revision
}
