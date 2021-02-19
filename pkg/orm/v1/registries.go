package v1

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
	"github.com/wujie1993/waves/pkg/util"
)

// AppRegistry 应用存储器
// +namespaced=true
type AppRegistry struct {
	registry.Registry
}

// appMutate 自定义应用内容写入填充逻辑
func appMutate(obj core.ApiObject) error {
	app := obj.(*App)
	if app.Spec.Platform == "" && len(app.Spec.Versions) > 0 {
		app.Spec.Platform = app.Spec.Versions[0].Platform
	}
	for index, versionApp := range app.Spec.Versions {
		if versionApp.LivenessProbe.InitialDelaySeconds < 0 {
			app.Spec.Versions[index].LivenessProbe.InitialDelaySeconds = 10
		}
		if versionApp.LivenessProbe.PeriodSeconds < 30 {
			app.Spec.Versions[index].LivenessProbe.PeriodSeconds = 60
		}
		if versionApp.LivenessProbe.TimeoutSeconds < 30 {
			app.Spec.Versions[index].LivenessProbe.TimeoutSeconds = 60
		}
	}
	return nil
}

// NewAppRegistry 实例化应用存储器
func NewAppRegistry() *AppRegistry {
	app := &AppRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindApp), true),
	}
	app.SetDefaultFinalizers([]string{
		core.FinalizerCleanRefConfigMap,
	})
	app.SetMutateHook(appMutate)
	return app
}

// AppInstanceRegistry 应用实例存储器
// +namespaced=true
type AppInstanceRegistry struct {
	registry.Registry
}

// appInstancePostCreate 自定义应用实例创建后逻辑
func appInstancePostCreate(obj core.ApiObject) error {
	hostRegistry := NewHostRegistry()

	appInstance := obj.(*AppInstance)

	if appInstance.Spec.Category == core.AppCategoryHostPlugin && len(appInstance.Spec.Modules) > 0 {
		// 只针对 第一个模块取主机
		for _, hostRef := range appInstance.Spec.Modules[0].HostRefs {
			// 获取插件关联的主机
			hostObj, err := hostRegistry.Get(context.TODO(), core.DefaultNamespace, hostRef)
			if err != nil {
				log.Error(err)
				return err
			}
			if hostObj == nil {
				err := e.Errorf("host %s not found", hostRef)
				log.Error(err)
				return err
			}
			host := hostObj.(*Host)

			host.Spec.Plugins = append(host.Spec.Plugins, HostPlugin{
				AppInstanceRef: AppInstanceRef{
					Namespace: appInstance.Metadata.Namespace,
					Name:      appInstance.Metadata.Name,
				},
				AppRef: appInstance.Spec.AppRef,
			})

			if _, err := hostRegistry.Update(context.TODO(), host, core.WithAllFields()); err != nil {
				log.Error(err)
				return err
			}
		}
	} else if appInstance.Spec.Category == core.AppCategoryAlgorithmPlugin && len(appInstance.Spec.Modules) > 0 {
		//更新host结构
		for _, hostRef := range appInstance.Spec.Modules[0].HostRefs {
			// 获取插件关联的主机
			hostObj, err := hostRegistry.Get(context.TODO(), core.DefaultNamespace, hostRef)
			if err != nil {
				log.Error(err)
			}
			if hostObj == nil {
				err := e.Errorf("host %s not found", hostRef)
				log.Error(err)
			}
			host := hostObj.(*Host)

			host.Spec.Sdks = append(host.Spec.Sdks, SdkPlugin{
				AppInstanceRef: AppInstanceRef{
					Namespace: appInstance.Metadata.Namespace,
					Name:      appInstance.Metadata.Name,
				},
				AppRef: AppRef{
					Name:    appInstance.Spec.AppRef.Name,
					Version: appInstance.Spec.AppRef.Version,
					//是否需要加入SupportMediaTypes
				},
			})

			if _, err := hostRegistry.Update(context.TODO(), host, core.WithAllFields()); err != nil {
				log.Error(err)
			}
		}
	}

	return nil
}

// appInstancePostUpdate 自定义应用实例更新前逻辑
func appInstancePreUpdate(obj core.ApiObject) error {
	appInstanceRegistry := NewAppInstanceRegistry()

	appInstance := obj.(*AppInstance)

	// 获取更新前的应用实例
	oldObj, err := appInstanceRegistry.Get(context.TODO(), appInstance.Metadata.Namespace, appInstance.Metadata.Name)
	if err != nil {
		log.Error(err)
		return err
	} else if oldObj == nil {
		return nil
	}
	oldAppInstance := oldObj.(*AppInstance)

	appRegistry := NewAppRegistry()
	// 获取关联的应用
	appObj, err := appRegistry.Get(context.TODO(), core.DefaultNamespace, appInstance.Spec.AppRef.Name)
	if err != nil {
		log.Error(err)
		return err
	}
	if appObj == nil {
		return nil
	}
	app := appObj.(*App)

	switch appInstance.Spec.Action {
	case core.AppActionConfigure:
		// 非Installed状态禁止配置操作
		if oldAppInstance.Status.Phase != core.PhaseInstalled {
			err := e.Errorf("not allow to configure when status phase not Installed")
			log.Error(err)
			return err
		}

		appInstance.Spec.Action = ""
		if oldAppInstance.SpecHash() == appInstance.SpecHash() {
			appInstance.Spec.Action = ""
		} else {
			appInstance.Spec.Action = core.AppActionConfigure
		}

		// 重置关联应用
		appInstance.Spec.AppRef = oldAppInstance.Spec.AppRef

		// 重置禁止修改的模块参数
		for _, versionApp := range app.Spec.Versions {
			for _, module := range versionApp.Modules {
				for _, arg := range module.Args {
					// 重置只读参数项
					if arg.Readonly {
						appInstance.Spec.SetModuleArgValue(module.Name, arg.Name, arg.Default)
					}
					// 重置禁止修改参数项
					if !arg.Modifiable {
						value, exist := oldAppInstance.Spec.GetModuleArgValue(module.Name, arg.Name)
						if exist {
							appInstance.Spec.SetModuleArgValue(module.Name, arg.Name, value)
						}
					}
				}
			}
		}
		// 重置禁止修改的全局参数
		for _, versionApp := range app.Spec.Versions {
			for _, arg := range versionApp.Global.Args {
				// 重置只读参数项
				if arg.Readonly {
					appInstance.Spec.SetGlobalArgValue(arg.Name, arg.Default)
				}
				// 重置禁止修改参数项
				if !arg.Modifiable {
					value, exist := oldAppInstance.Spec.GetGlobalArgValue(arg.Name)
					if exist {
						appInstance.Spec.SetGlobalArgValue(arg.Name, value)
					}
				}
			}
		}
	case core.AppActionUpgrade:
		oldSpec, err := oldObj.SpecEncode()
		if err != nil {
			return err
		}
		appInstance.Metadata.Annotations[core.AnnotationPrefix+"upgrade/last-applied-configuration"] = string(oldSpec)
	}

	return nil
}

// appInstanceValidate 自定义应用实例内容写入校验逻辑
func appInstanceValidate(obj core.ApiObject) error {
	hostRegistry := NewHostRegistry()
	appRegistry := NewAppRegistry()

	appInstance := obj.(*AppInstance)

	// 验证应用是否存在
	appObj, err := appRegistry.Get(context.TODO(), core.DefaultNamespace, appInstance.Spec.AppRef.Name)
	if err != nil {
		return err
	} else if appObj == nil {
		return e.Errorf("referred app %s not found", appInstance.Spec.AppRef.Name)
	}
	app := appObj.(*App)

	// 验证应用版本是否存在
	var appVersionExist bool
	for _, versionApp := range app.Spec.Versions {
		if versionApp.Version == appInstance.Spec.AppRef.Version {
			appVersionExist = true
			break
		}
	}
	if !appVersionExist {
		return e.Errorf("referred app version %s not found", appInstance.Spec.AppRef.Version)
	}

	switch app.Spec.Category {
	case core.AppCategoryHostPlugin:
		if len(appInstance.Spec.Modules) > 0 {
			// 限制主机插件在一台主机上只能安装一个
			for _, hostRef := range appInstance.Spec.Modules[0].HostRefs {
				// 获取插件关联的主机
				hostObj, err := hostRegistry.Get(context.TODO(), core.DefaultNamespace, hostRef)
				if err != nil {
					log.Error(err)
					return err
				}
				if hostObj == nil {
					err := e.Errorf("host %s not found", hostRef)
					log.Error(err)
					return err
				}
				host := hostObj.(*Host)

				for _, plugin := range host.Spec.Plugins {
					// 判断插件是否已经存在
					if plugin.AppRef.Name == appInstance.Spec.AppRef.Name && (plugin.AppInstanceRef.Name != appInstance.Metadata.Name || plugin.AppInstanceRef.Namespace != appInstance.Metadata.Namespace) {
						err := e.Errorf("plugin %s already exist in host %s", plugin.AppRef.Name, hostRef)
						log.Error(err)
						return err
					}
				}
			}
		}
	case core.AppCategoryAlgorithmPlugin:
		return errors.New("algorithm plugin is not allow to create app instance")
	}

	return nil
}

// appInstanceMutate 自定义应用实例内容写入填充逻辑
func appInstanceMutate(obj core.ApiObject) error {
	appInstance := obj.(*AppInstance)

	appRegistry := NewAppRegistry()
	// 获取关联的应用
	appObj, err := appRegistry.Get(context.TODO(), core.DefaultNamespace, appInstance.Spec.AppRef.Name)
	if err != nil {
		log.Error(err)
		return err
	}
	if appObj == nil {
		return nil
	}
	app := appObj.(*App)

	// 填充应用实例分类
	appInstance.Spec.Category = app.Spec.Category

	// 填充健康检查配置项
	if appInstance.Spec.LivenessProbe.InitialDelaySeconds < 0 {
		appInstance.Spec.LivenessProbe.InitialDelaySeconds = 10
	}
	if appInstance.Spec.LivenessProbe.PeriodSeconds < 60 {
		appInstance.Spec.LivenessProbe.PeriodSeconds = 60
	}
	if appInstance.Spec.LivenessProbe.TimeoutSeconds < 60 {
		appInstance.Spec.LivenessProbe.TimeoutSeconds = 60
	}

	cmRegistry := NewConfigMapRegistry()
	for moduleIndex, module := range appInstance.Spec.Modules {
		// 无需存储Notes字段
		appInstance.Spec.Modules[moduleIndex].Notes = ""

		// 填充配置文件哈希值，确保配置文件更新时，应用实例内容也发生更新
		if module.ConfigMapRef.Name != "" {
			cm, err := cmRegistry.Get(context.TODO(), module.ConfigMapRef.Namespace, module.ConfigMapRef.Name)
			if err != nil {
				log.Error(err)
				return err
			}

			if cm != nil {
				appInstance.Spec.Modules[moduleIndex].ConfigMapRef.Hash = cm.SpecHash()
				appInstance.Spec.Modules[moduleIndex].ConfigMapRef.Revision = cm.GetMetadata().ResourceVersion
			}
		}
		if module.AdditionalConfigs.ConfigMapRef.Name != "" {
			cm, err := cmRegistry.Get(context.TODO(), module.AdditionalConfigs.ConfigMapRef.Namespace, module.AdditionalConfigs.ConfigMapRef.Name)
			if err != nil {
				log.Error(err)
				return err
			}

			if cm != nil {
				appInstance.Spec.Modules[moduleIndex].AdditionalConfigs.ConfigMapRef.Hash = cm.SpecHash()
			}
		}

		// 在安装或升级应用实例时，更新模块版本
		switch appInstance.Spec.Action {
		case core.AppActionInstall:
			appInstance.Spec.Modules[moduleIndex].AppVersion = appInstance.Spec.AppRef.Version
		}

		// 模块版本为空时，使用当前应用版本填充
		if appInstance.Spec.Modules[moduleIndex].AppVersion == "" {
			appInstance.Spec.Modules[moduleIndex].AppVersion = appInstance.Spec.AppRef.Version
		}
	}

	return nil
}

// appInstanceDecorate 自定义应用实例内容读取装饰逻辑
func appInstanceDecorate(obj core.ApiObject) error {
	appInstance := obj.(*AppInstance)

	// 渲染模块说明
	if appInstance.Status.Phase == core.PhaseInstalled {
		appRegistry := NewAppRegistry()

		// 获取关联的应用
		appObj, err := appRegistry.Get(context.TODO(), core.DefaultNamespace, appInstance.Spec.AppRef.Name)
		if err != nil {
			log.Error(err)
			return err
		}
		if appObj == nil {
			return nil
		}
		app := appObj.(*App)

		var versionApp AppVersion
		for _, version := range app.Spec.Versions {
			if version.Version == appInstance.Spec.AppRef.Version {
				versionApp = version
			}
		}

		for moduleIndex, module := range appInstance.Spec.Modules {
			for _, appModule := range versionApp.Modules {
				if module.Name != appModule.Name {
					continue
				}
				notes := make(map[string]interface{})
				// 填充Notes
				if versionApp.Platform == core.AppPlatformBareMetal {
					hostRegistry := NewHostRegistry()
					hosts := []string{}
					for _, hostRef := range module.HostRefs {
						hostObj, err := hostRegistry.Get(context.TODO(), "", hostRef)
						if err != nil {
							log.Error(err)
							return err
						} else if hostObj == nil {
							continue
						}
						host := hostObj.(*Host)
						hosts = append(hosts, host.Spec.SSH.Host)
					}
					notes["Hosts"] = hosts
				}
				args := make(map[string]interface{})
				for _, arg := range module.Args {
					args[arg.Name] = arg.Value
				}
				notes["Args"] = args

				tpl, err := template.New("notes").Parse(appModule.Notes)
				if err != nil {
					log.Error(err)
				} else {
					var buffer bytes.Buffer
					if err := tpl.Execute(&buffer, notes); err != nil {
						log.Error(err)
					}

					appInstance.Spec.Modules[moduleIndex].Notes = buffer.String()
				}
			}
		}
	}
	return nil
}

// NewAppInstanceRegistry 实例化应用实例存储器
func NewAppInstanceRegistry() *AppInstanceRegistry {
	r := &AppInstanceRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindAppInstance), true),
	}
	r.SetDefaultFinalizers([]string{
		core.FinalizerCleanRefEvent,
		core.FinalizerReleaseRefGPU,
		core.FinalizerCleanRefConfigMap,
	})
	r.SetValidateHook(appInstanceValidate)
	r.SetMutateHook(appInstanceMutate)
	r.SetDecorateHook(appInstanceDecorate)
	r.SetPreUpdateHook(appInstancePreUpdate)
	r.SetPostCreateHook(appInstancePostCreate)
	return r
}

// AuditRegistry 审计日志存储器
type AuditRegistry struct {
	registry.Registry
}

// Record 记录一条新的审计日志
func (r AuditRegistry) Record(audit *Audit) error {
	var namespaceMsg, shortNameMsg, actionMsg, descMsg string

	if audit.Spec.Msg != "" {
		descMsg = ", 备注: " + audit.Spec.Msg
	}

	if audit.Spec.ResourceRef.Namespace != "" {
		namespaceMsg = "在命名空间 " + audit.Spec.ResourceRef.Namespace + " 下"
	}

	if shortName := audit.Metadata.Annotations["ShortName"]; shortName != "" {
		shortNameMsg = "(" + shortName + ")"
	}

	switch audit.Spec.Action {
	case core.AuditActionCreate:
		actionMsg = "创建"
	case core.AuditActionUpdate:
		actionMsg = "更新"
	case core.AuditActionDelete:
		actionMsg = "删除"
	default:
		return errors.New("unsupport method")
	}

	audit.Metadata.Name = fmt.Sprintf("%d", time.Now().UnixNano())
	audit.Spec.Msg = namespaceMsg + actionMsg + core.GetKindMsg(audit.Spec.ResourceRef.Kind) + " " + audit.Spec.ResourceRef.Name + shortNameMsg + descMsg

	if _, err := r.Create(context.TODO(), audit); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// NewAuditRegistry 实例化审计日志存储器
func NewAuditRegistry() *AuditRegistry {
	return &AuditRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindAudit), false),
	}
}

// ConfigMapRegistry 配置字典存储器
// +namespaced=true
type ConfigMapRegistry struct {
	registry.Registry
}

// NewConfigMapRegistry 实例化配置字典存储器
func NewConfigMapRegistry() *ConfigMapRegistry {
	r := &ConfigMapRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindConfigMap), true),
	}
	r.SetDefaultFinalizers([]string{
		core.FinalizerCleanRevision,
	})
	r.SetRevisioner(NewConfigMapRevision())
	return r
}

// EventRegistry 事件日志存储器
type EventRegistry struct {
	registry.Registry
}

// Record 记录一条新的事件日志
func (r EventRegistry) Record(event *Event) error {
	var shortNameMsg string
	if shortName := event.Metadata.Annotations["ShortName"]; shortName != "" {
		shortNameMsg = "(" + shortName + ")"
	}
	appendMsg := event.Spec.Msg
	phase := event.Status.Phase

	event.Metadata.Name = fmt.Sprintf("%d", time.Now().UnixNano())
	event.Spec.Msg = core.GetKindMsg(event.Spec.ResourceRef.Kind) + " " + event.Spec.ResourceRef.Name + shortNameMsg + " "
	switch event.Status.Phase {
	case core.PhaseCompleted:
		event.Spec.Msg += core.GetActionMsg(event.Spec.Action) + "完成"
	case core.PhaseFailed:
		event.Spec.Msg += core.GetActionMsg(event.Spec.Action) + "失败"
	default:
		event.Status.Phase = core.PhaseWaiting
		event.Spec.Msg += "开始" + core.GetActionMsg(event.Spec.Action)
	}

	if appendMsg != "" {
		event.Spec.Msg += "，备注：" + appendMsg
	}

	if _, err := r.Create(context.TODO(), event); err != nil {
		log.Error(err)
		return err
	}
	if phase != core.PhaseWaiting {
		if _, err := r.UpdateStatusPhase(event.Metadata.Namespace, event.Metadata.Name, phase); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

// NewEventRegistry 实例化事件日志存储器
func NewEventRegistry() *EventRegistry {
	r := &EventRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindEvent), false),
	}
	r.SetDefaultFinalizers([]string{
		core.FinalizerCleanRefJob,
	})
	return r
}

// GPURegistry 显卡存储器
type GPURegistry struct {
	registry.Registry
}

// GetGPUName 获取显卡标识名
func (r GPURegistry) GetGPUName(hostRef string, slot int) string {
	return fmt.Sprintf("%s-slot-%d", hostRef, slot)
}

// NewGPURegistry 实例化显卡存储器
func NewGPURegistry() *GPURegistry {
	return &GPURegistry{
		Registry: registry.NewRegistry(newGVK(core.KindGPU), false),
	}
}

// HostRegistry 主机存储器
type HostRegistry struct {
	registry.Registry
}

// NewHostRegistry 实例化主机存储器
func NewHostRegistry() *HostRegistry {
	r := &HostRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindHost), false),
	}
	r.SetDefaultFinalizers([]string{
		core.FinalizerCleanRefGPU,
		core.FinalizerCleanRefEvent,
	})
	return r
}

// JobRegistry 任务存储器
type JobRegistry struct {
	registry.Registry
}

// GetLogPath 获取任务工作目录路径
func (r JobRegistry) GetLogPath(jobsDir string, jobName string) (string, error) {
	obj, err := r.Get(context.TODO(), "", jobName)
	if err != nil {
		return "", err
	}
	if obj == nil {
		return "", nil
	}
	job := obj.(*Job)
	return path.Join(jobsDir, job.Metadata.Uid, "ansible.log"), nil
}

// GetLog 获取任务的运行日志
func (r JobRegistry) GetLog(jobsDir string, jobName string) ([]byte, error) {
	jobPath, err := r.GetLogPath(jobsDir, jobName)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(jobPath)
}

// WatchLog 监听任务运行日志的输出内容
func (r JobRegistry) WatchLog(ctx context.Context, jobsDir string, jobName string) (<-chan string, error) {
	jobPath, err := r.GetLogPath(jobsDir, jobName)
	if err != nil {
		return nil, err
	}
	return util.Tailf(ctx, jobPath)
}

// NewJobRegistry 实例化任务存储器
func NewJobRegistry() *JobRegistry {
	r := &JobRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindJob), false),
	}
	r.SetDefaultFinalizers([]string{
		core.FinalizerCleanRefConfigMap,
		core.FinalizerCleanJobWorkDir,
	})
	return r
}

// K8sConfigRegistry K8S集群存储器
// +namespaced=true
type K8sConfigRegistry struct {
	registry.Registry
}

func (r K8sConfigRegistry) GetFirstMasterHost(name string) (*Host, error) {
	// 获取k8s集群节点
	k8sObj, err := r.Get(context.TODO(), core.DefaultNamespace, name)
	if err != nil {
		return nil, err
	} else if k8sObj == nil {
		return nil, e.Errorf("k8s cluster '%s' not found", name)
	}
	k8s := k8sObj.(*K8sConfig)

	if len(k8s.Spec.K8SMaster.Hosts) < 1 {
		return nil, e.Errorf("k8s cluster '%s' does not have any master hosts exist", name)
	}
	// 获取k8s集群第一个master节点
	hostRef := k8s.Spec.K8SMaster.Hosts[0].ValueFrom.HostRef
	hostRegistry := NewHostRegistry()
	hostObj, err := hostRegistry.Get(context.TODO(), "", hostRef)
	if err != nil {
		return nil, err
	} else if hostObj == nil {
		return nil, e.Errorf("host %s not found", hostRef)
	}
	return hostObj.(*Host), nil
}

// NewK8sConfigRegistry 实例化K8S集群存储器
func NewK8sConfigRegistry() *K8sConfigRegistry {
	r := &K8sConfigRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindK8sConfig), true),
	}
	r.SetDefaultFinalizers([]string{
		core.FinalizerCleanRefEvent,
	})
	return r
}

// NamespaceRegistry 命名空间存储器
type NamespaceRegistry struct {
	registry.Registry
}

// NewNamespaceRegistry 实例化命名空间存储器
func NewNamespaceRegistry() *NamespaceRegistry {
	r := &NamespaceRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindNamespace), false),
	}
	r.SetDefaultFinalizers([]string{
		core.FinalizerCleanRefConfigMap,
	})
	return r
}

// PkgRegistry 部署包存储器
type PkgRegistry struct {
	registry.Registry
}

// NewPkgRegistry 实例化部署包存储器
func NewPkgRegistry() *PkgRegistry {
	return &PkgRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindPkg), false),
	}
}

// ProjectRegistry 项目存储器
type ProjectRegistry struct {
	registry.Registry
}

// projectMutate 自定义项目内容写入填充逻辑
func projectMutate(obj core.ApiObject) error {
	project := obj.(*Project)

	// 如果项目没有关联同名的命名空间，则创建并关联命名空间
	nsExist := false
	for _, referNs := range project.ReferNamespaces {
		if referNs == project.Metadata.Name {
			nsExist = true
		}
	}
	if !nsExist {
		nsRegistry := NewNamespaceRegistry()
		nsObj, err := nsRegistry.Get(context.TODO(), "", project.Metadata.Name)
		if err != nil {
			return err
		}
		if nsObj == nil {
			ns := NewNamespace()
			ns.Metadata.Name = project.Metadata.Name
			if _, err := nsRegistry.Create(context.TODO(), ns); err != nil {
				return err
			}
		}
		project.ReferNamespaces = append(project.ReferNamespaces, project.Metadata.Name)
	}
	return nil
}

// NewProjectRegistry 实例化项目存储器
func NewProjectRegistry() *ProjectRegistry {
	r := &ProjectRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindProject), false),
	}
	r.SetMutateHook(projectMutate)
	return r
}

// RevisionRegistry 修订版本存储器
type RevisionRegistry struct {
	registry.Registry
}

// NewRevisionRegistry 实例化修订版本存储器
func NewRevisionRegistry() *RevisionRegistry {
	r := &RevisionRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindRevision), false),
	}
	return r
}
