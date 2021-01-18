package core

import (
	"encoding/json"
	"errors"
)

const (
	AnnotationPrefix                   = "pcitech.io/"
	AnnotationJobPrefix                = AnnotationPrefix + "job/"
	AnnotationAlgorithmPluginPrefix    = AnnotationPrefix + "algorithm-plugin/"
	AnnotationLastAppliedConfiguration = AnnotationPrefix + "last-applied-configuration"

	Group        = "core"
	ApiVersionV1 = "v1"

	AppCategoryCustomize       = "customize"
	AppCategoryThirdParty      = "thirdParty"
	AppCategoryHostPlugin      = "hostPlugin"
	AppCategoryAlgorithmPlugin = "algorithmPlugin"

	AppPlatformBareMetal = "bareMetal"
	AppPlatformK8s       = "k8s"

	AppActionInstall     = "install"
	AppActionConfigure   = "configure"
	AppActionHealthcheck = "healthcheck"
	AppActionUninstall   = "uninstall"
	AppActionUpgrade     = "upgrade"
	AppActionRevert      = "revert"

	AuditActionCreate = "create"
	AuditActionUpdate = "update"
	AuditActionDelete = "delete"

	DefaultEtcdEndpoint = "localhost:2379"

	DefaultNamespace = "default"

	RegistryPrefix = "/registry"

	KindApp         = "app"
	KindAppInstance = "appInstance"
	KindAudit       = "audit"
	KindEvent       = "event"
	KindHost        = "host"
	KindJob         = "job"
	KindConfigMap   = "configMap"
	KindK8sConfig   = "k8sconfig"
	KindK8sLabel    = "k8slabel"
	KindNamespace   = "namespace"
	KindPkg         = "pkg"
	KindGPU         = "gpu"
	KindProject     = "project"
	KindRevision    = "revision"

	ConditionTypeConnected   = "Connected"
	ConditionTypeInitialized = "Initialized"
	ConditionTypeInstalled   = "Installed"
	ConditionTypeHealthy     = "Healthy"
	ConditionTypeReady       = "Ready"
	ConditionTypeConfigured  = "Configured"
	ConditionTypeRun         = "Run"

	ConditionStatusTrue  = "True"
	ConditionStatusFalse = "False"

	EventActionInstall       = "Install"
	EventActionConfigure     = "Configure"
	EventActionUninstall     = "Uninstall"
	EventActionHealthCheck   = "HealthCheck"
	EventActionUpgrade       = "Upgrade"
	EventActionRevert        = "Revert"
	EventActionLabel         = "Label"
	EventActionUnLabel       = "UnLabel"
	EventActionConnect       = "Connect"
	EventActionInitial       = "Initial"
	EventActionUninstallNode = "UninstallNode"
	EventActionInstallNode   = "InstallNode"

	FinalizerCleanRefJob       = "CleanRefJob"
	FinalizerCleanJobWorkDir   = "CleanJobWorkDir"
	FinalizerCleanRefGPU       = "CleanRefGPU"
	FinalizerReleaseRefGPU     = "ReleaseRefGPU"
	FinalizerCleanRefEvent     = "CleanRefEvent"
	FinalizerCleanRefConfigMap = "CleanRefConfigMap"
	FinalizerCleanRevision     = "CleanRevision"

	JobExecTypeAnsible         = "ansible"
	JobDefaultFailureThreshold = 1
	JobDefaultTimeoutSeconds   = 3600

	PhaseRunning       = "Running"
	PhaseInitialing    = "Initialing"
	PhaseInstalling    = "Installing"
	PhaseUninstalling  = "Uninstalling"
	PhaseWaiting       = "Waiting"
	PhaseFailed        = "Failed"
	PhaseDeleting      = "Deleting"
	PhaseBound         = "Bound"
	PhaseCompleted     = "Completed"
	PhaseConfiguring   = "Configuring"
	PhaseUpgradeing    = "Upgrading"
	PhaseReverting     = "Reverting"
	PhaseConnecting    = "Connecting"
	PhaseCrashing      = "Crashing"
	PhaseReady         = "Ready"
	PhaseNotReady      = "NotReady"
	PhaseInstalled     = "Installed"
	PhaseUninstalled   = "Uninstalled"
	PhaseLabel         = "Label"
	PhaseUnLabel       = "UnLabel"
	PhaseUninstallNode = "UninstallNode"
	PhaseInCompleted   = "InstallCompleted"
	PhaseUnCompleted   = "UninstallCompleted"

	PkgProvisionFull = "full"
	PkgProvisionThin = "thin"

	ValidNameRegex = `^[a-zA-Z0-9_\-\.]{1,256}$`
)

var (
	// 已注册的资源类型列表
	kinds = []string{
		KindApp,
		KindAppInstance,
		KindAudit,
		KindConfigMap,
		KindEvent,
		KindGPU,
		KindHost,
		KindJob,
		KindK8sConfig,
		KindPkg,
		KindProject,
	}

	// 资源复数别名，用于接口url，如：api/<version>/<plural>
	kindPluralMap = map[string]string{
		KindApp:         "apps",
		KindAppInstance: "appinstances",
		KindAudit:       "audits",
		KindConfigMap:   "configmaps",
		KindEvent:       "events",
		KindGPU:         "gpus",
		KindHost:        "hosts",
		KindJob:         "jobs",
		KindK8sConfig:   "k8sconfig",
		KindPkg:         "pkgs",
		KindProject:     "projects",
	}

	// 资源单数名称
	kindSingularMap = map[string]string{
		KindApp:         "app",
		KindAppInstance: "appinstance",
		KindAudit:       "audit",
		KindConfigMap:   "configmap",
		KindEvent:       "event",
		KindGPU:         "gpu",
		KindHost:        "host",
		KindJob:         "job",
		KindK8sConfig:   "k8sconfig",
		KindPkg:         "pkg",
		KindProject:     "project",
	}

	// 资源简称，便于命令行使用资源
	kindShortNamesMap = map[string][]string{
		KindAppInstance: []string{"ins"},
		KindConfigMap:   []string{"cm"},
		KindHost:        []string{"node", "nodes"},
		KindK8sConfig:   []string{"k8s"},
	}

	// 资源类型描述
	kindMsg = map[string]string{
		KindApp:         "应用",
		KindAppInstance: "应用实例",
		KindAudit:       "审计",
		KindConfigMap:   "配置",
		KindProject:     "项目名称",
		KindHost:        "主机",
		KindJob:         "任务",
		KindK8sConfig:   "K8s集群",
		KindPkg:         "部署包",
		KindGPU:         "显卡",
	}

	// 操作行为描述
	actionMsg = map[string]string{
		EventActionUpgrade:       "版本升级",
		EventActionRevert:        "版本回退",
		EventActionConfigure:     "配置",
		EventActionConnect:       "连接",
		EventActionHealthCheck:   "健康检查",
		EventActionInitial:       "初始化",
		EventActionInstall:       "安装",
		EventActionLabel:         "打标签",
		EventActionUnLabel:       "删除标签",
		EventActionUninstall:     "卸载",
		EventActionUninstallNode: "卸载节点",
		EventActionInstallNode:   "新增节点",
	}

	// 应用分类描述
	categoryMsg = map[string]string{
		AppCategoryAlgorithmPlugin: "算法插件",
		AppCategoryCustomize:       "业务应用",
		AppCategoryHostPlugin:      "主机插件",
		AppCategoryThirdParty:      "基础组件",
	}
)

type ValueFrom struct {
	ConfigMapKeyRef ConfigMapKeyRef
}

type ConfigMapKeyRef struct {
	Name string
	Key  string
}

// DeepCopy 深度复制两个对象，src和dest必须是引用传递类型如指针，切片或字典
func DeepCopy(src, dest interface{}) error {
	if dest == nil {
		return errors.New("nil pointer of dest obj")
	}
	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, dest)
}

// GetKindMsg 根据资源类型获取其简称，在资源类型不存在的情况下返回类型描述
func GetKindMsg(kind string) string {
	msg, ok := kindMsg[kind]
	if !ok {
		return kind
	}
	return msg
}

// GetActionMsg 根据行为类型获取其简称，在行为类型不存在的情况下返回行为描述
func GetActionMsg(action string) string {
	msg, ok := actionMsg[action]
	if !ok {
		return action
	}
	return msg
}

// GetCategoryMsg 根据应用分类获取其简称
func GetCategoryMsg(category string) string {
	msg, ok := categoryMsg[category]
	if !ok {
		return category
	}
	return msg
}

// SearchKind 根据字符串查询资源类型
func SearchKind(keyword string) string {
	// 根据类型列表查询
	for _, kind := range kinds {
		if keyword == kind {
			return kind
		}
	}

	// 根据单数别名查询
	for kind, singular := range kindSingularMap {
		if singular == keyword {
			return kind
		}
	}

	// 根据复数别名查询
	for kind, plural := range kindPluralMap {
		if plural == keyword {
			return kind
		}
	}

	// 根据简称查询
	for kind, shortNames := range kindShortNamesMap {
		for _, shortName := range shortNames {
			if shortName == keyword {
				return kind
			}
		}
	}
	return ""
}

func GetPlural(kind string) string {
	plural, ok := kindPluralMap[kind]
	if ok {
		return plural
	}
	return ""
}

func GetSingular(kind string) string {
	singular, ok := kindSingularMap[kind]
	if ok {
		return singular
	}
	return ""
}
