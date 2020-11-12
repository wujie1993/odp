package core

import (
	"encoding/json"
	"errors"
)

const (
	AnnotationPrefix                = "pcitech.io/"
	AnnotationJobPrefix             = AnnotationPrefix + "job/"
	AnnotationAlgorithmPluginPrefix = AnnotationPrefix + "algorithm-plugin/"

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

	AuditActionCreate = "create"
	AuditActionUpdate = "update"
	AuditActionDelete = "delete"

	DefaultEtcdEndpoint = "localhost:2379"

	DefaultNamespace = "default"

	RegistryPrefix = "/prophet"

	KindApp         = "app"
	KindAppInstance = "appInstance"
	KindAudit       = "audit"
	KindEvent       = "event"
	KindHost        = "host"
	KindJob         = "job"
	KindConfigMap   = "configMap"
	KindK8sConfig   = "k8sconfig"
	KindK8sLabel    = "k8slabel"
	KindPkg         = "pkg"
	KindGPU         = "gpu"
	KindProject     = "project"

	ConditionTypeConnected   = "Connected"
	ConditionTypeInitialized = "Initialized"
	ConditionTypeInstalled   = "Installed"
	ConditionTypeHealthy     = "Healthy"
	ConditionTypeReady       = "Ready"
	ConditionTypeConfigured  = "Configured"
	ConditionTypeRun         = "Run"

	ConditionStatusTrue  = "True"
	ConditionStatusFalse = "False"

	EventActionInstall        = "Install"
	EventActionConfigure      = "Configure"
	EventActionUninstall      = "Uninstall"
	EventActionHealthCheck    = "HealthCheck"
	EventActionUpgrade        = "Upgrade"
	EventActionUpgradeBackoff = "UpgradeBackoff"
	EventActionLabel          = "Label"
	EventActionUnLabel        = "UnLabel"
	EventActionConnect        = "Connect"
	EventActionInitial        = "Initial"
	EventActionUninstallNode  = "UninstallNode"
	EventActionInstallNode    = "InstallNode"

	FinalizerCleanRefJob       = "CleanRefJob"
	FinalizerCleanJobWorkDir   = "CleanJobWorkDir"
	FinalizerCleanRefGPU       = "CleanRefGPU"
	FinalizerReleaseRefGPU     = "ReleaseRefGPU"
	FinalizerCleanRefEvent     = "CleanRefEvent"
	FinalizerCleanRefConfigMap = "CleanRefConfigMap"

	JobExecTypeAnsible         = "ansible"
	JobDefaultFailureThreshold = 1
	JobDefaultTimeoutSeconds   = 3600

	PhaseRunning           = "Running"
	PhaseInitialing        = "Initialing"
	PhaseInstalling        = "Installing"
	PhaseUninstalling      = "Uninstalling"
	PhaseWaiting           = "Waiting"
	PhaseFailed            = "Failed"
	PhaseDeleting          = "Deleting"
	PhaseBound             = "Bound"
	PhaseCompleted         = "Completed"
	PhaseConfiguring       = "Configuring"
	PhaseUpgradeing        = "Upgrading"
	PhaseUpgradeBackoffing = "UpgradeBackoffing"
	PhaseConnecting        = "Connecting"
	PhaseCrashing          = "Crashing"
	PhaseReady             = "Ready"
	PhaseNotReady          = "NotReady"
	PhaseInstalled         = "Installed"
	PhaseUninstalled       = "Uninstalled"
	PhaseLabel             = "Label"
	PhaseUnLabel           = "UnLabel"
	PhaseUninstallNode     = "UninstallNode"
	PhaseInCompleted       = "InstallCompleted"
	PhaseUnCompleted       = "UninstallCompleted"

	PkgProvisionFull = "full"
	PkgProvisionThin = "thin"

	ValidNameRegex = `^[a-zA-Z0-9_\-\.]{1,256}$`
)

var kindMsg map[string]string
var actionMsg map[string]string

func init() {
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
	actionMsg = map[string]string{
		EventActionUpgrade:        "版本升级",
		EventActionUpgradeBackoff: "版本回退",
		EventActionConfigure:      "配置",
		EventActionConnect:        "连接",
		EventActionHealthCheck:    "健康检查",
		EventActionInitial:        "初始化",
		EventActionInstall:        "安装",
		EventActionLabel:          "打标签",
		EventActionUnLabel:        "删除标签",
		EventActionUninstall:      "卸载",
		EventActionUninstallNode:  "卸载节点",
		EventActionInstallNode:    "新增节点",
	}
}

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
