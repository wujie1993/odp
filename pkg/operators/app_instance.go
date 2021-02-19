package operators

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/wujie1993/waves/pkg/ansible"
	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/orm/v2"
	"github.com/wujie1993/waves/pkg/setting"
)

const (
	ArgTypeInteger = "integer"
	ArgTypeNumber  = "number"
	ArgTypeString  = "string"
	ArgTypeBoolean = "boolean"

	ArgFormatInt32     = "int32"
	ArgFormatInt64     = "int64"
	ArgFormatPort      = "port"
	ArgFormatFloat     = "float"
	ArgFormatDouble    = "double"
	ArgFormatDate      = "date"
	ArgFormatPassword  = "password"
	ArgFormatArray     = "array"
	ArgFormatGroupHost = "groupHost"
)

// ModuleAction 描述应用实例模块中每个主机所需要执行的操作，用于生成任务
type ModuleAction struct {
	ModuleName    string
	AppVersion    string
	ReplicaIndex  int
	HostActionMap map[string]string
}

// AppInstanceOperator 应用实例控制器用于处理应用实例的添加，更新和删除行为, 每个应用实例都会与一个应用版本绑定, 根据操作行为的不同, 会创建对应的任务进行处理.
type AppInstanceOperator struct {
	BaseOperator

	healthchecks *HealthChecks
	revisioner   registry.Revisioner
}

// handleAppInstance 处理应用实例的变更操作
func (o *AppInstanceOperator) handleAppInstance(ctx context.Context, obj core.ApiObject) error {
	appInstance := obj.(*v2.AppInstance)
	log.Infof("'%s' is %s", appInstance.GetKey(), appInstance.Status.Phase)

	o.setHealthCheck(ctx, obj)

	// 根据应用实例的状态做对应的处理
	switch appInstance.Status.Phase {
	case core.PhaseWaiting:
		// 处于等待中状态, 根据应用实例的操作行为创建对应的任务, 并绑定到应用实例上

		// 忽略内容体没有发生更新的应用实例
		if hash, ok := o.applyings.Get(appInstance.GetKey()); ok && hash == appInstance.SpecHash() {
			return nil
		}

		// 填充事件信息与应用实例状态，忽略没有赋予合法操作行为的应用实例
		switch appInstance.Spec.Action {
		case core.AppActionInstall:
			appInstance.Status.Phase = core.PhaseInstalling
		case core.AppActionUninstall:
			appInstance.Status.Phase = core.PhaseUninstalling
		case core.AppActionConfigure:
			appInstance.Status.Phase = core.PhaseConfiguring
		case core.AppActionUpgrade:
			appInstance.Status.Phase = core.PhaseUpgrading
		case core.AppActionRevert:
			appInstance.Status.Phase = core.PhaseReverting
		default:
			return nil
		}

		// 记录应用实例的内容哈希值, 用于后续比较内容体是否有更新，计算哈希值时忽略Action字段
		appInstance.Spec.Action = ""
		o.applyings.Set(appInstance.GetKey(), appInstance.SpecHash())

		// 等待处理的应用实例健康状态会被重置
		appInstance.Status.UnsetCondition(core.ConditionTypeHealthy)
		if _, err := o.helper.V2.AppInstance.Update(context.TODO(), appInstance, core.WithAllFields()); err != nil {
			log.Error(err)
			return err
		}
	case core.PhaseUninstalling:
		o.uninstallAppInstance(ctx, appInstance)
	case core.PhaseConfiguring:
		o.updateAppInstance(ctx, obj, core.EventActionConfigure)
	case core.PhaseInstalling:
		o.installAppInstance(ctx, appInstance)
	case core.PhaseUpgrading:
		o.updateAppInstance(ctx, obj, core.EventActionUpgrade)
	case core.PhaseReverting:
		o.updateAppInstance(ctx, obj, core.EventActionRevert)
	case core.PhaseDeleting:
		o.delete(ctx, obj)
	}
	return nil
}

// finalizeAppInstance 清除应用实例的关联资源
func (o AppInstanceOperator) finalizeAppInstance(ctx context.Context, obj core.ApiObject) error {
	appInstance := obj.(*v2.AppInstance)
	// 每次只处理一项Finalizer
	switch appInstance.Metadata.Finalizers[0] {
	case core.FinalizerCleanRefEvent:
		// 同步删除关联的事件
		eventList, err := o.helper.V1.Event.List(context.TODO(), "")
		if err != nil {
			log.Error(err)
			return err
		}
		for _, eventObj := range eventList {
			event := eventObj.(*v1.Event)
			if event.Spec.ResourceRef.Kind == core.KindAppInstance && event.Spec.ResourceRef.Namespace == appInstance.Metadata.Namespace && event.Spec.ResourceRef.Name == appInstance.Metadata.Name {
				if _, err := o.helper.V1.Event.Delete(context.TODO(), "", event.Metadata.Name); err != nil {
					log.Error(err)
					return err
				}
			}
		}
	case core.FinalizerCleanRefConfigMap:
		// 清除关联的配置字典
		for _, module := range appInstance.Spec.Modules {
			for _, replica := range module.Replicas {
				if replica.ConfigMapRef.Name != "" {
					configMapDeleteCtx, _ := context.WithTimeout(ctx, time.Second*5)
					if _, err := o.helper.V1.ConfigMap.Delete(configMapDeleteCtx, replica.ConfigMapRef.Namespace, replica.ConfigMapRef.Name); err != nil {
						log.Error(err)
						return err
					}
				}
			}
		}
		if appInstance.Spec.Global.ConfigMapRef.Name != "" {
			if _, err := o.helper.V1.ConfigMap.Delete(context.TODO(), appInstance.Spec.Global.ConfigMapRef.Namespace, appInstance.Spec.Global.ConfigMapRef.Name); err != nil {
				log.Error(err)
				return err
			}
		}
	case core.FinalizerReleaseRefGPU:
		// 释放关联的GPU资源
		if err := o.releaseGPU(appInstance); err != nil {
			log.Error(err)
			return err
		}
	case core.FinalizerCleanRevision:
		if err := o.revisioner.DeleteAllRevisions(ctx, appInstance.Metadata.Namespace, appInstance.Metadata.Name); err != nil {
			log.Error(err)
			return err
		}
	case core.FinalizerCleanHostPlugin:
		hostRegistry := v2.NewHostRegistry()

		if appInstance.Spec.Category == core.AppCategoryHostPlugin && len(appInstance.Spec.Modules) > 0 && len(appInstance.Spec.Modules[0].Replicas) > 0 {
			for _, hostRef := range appInstance.Spec.Modules[0].Replicas[0].HostRefs {
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
				host := hostObj.(*v2.Host)

				for index, plugin := range host.Info.Plugins {
					if plugin.AppRef.Name == appInstance.Spec.AppRef.Name && plugin.AppRef.Version == appInstance.Spec.AppRef.Version {
						host.Info.Plugins = append(host.Info.Plugins[:index], host.Info.Plugins[index+1:]...)
					}
				}

				if _, err := hostRegistry.Update(context.TODO(), host, core.WithAllFields()); err != nil {
					log.Error(err)
					return err
				}
			}
		}

		return nil
	}
	return nil
}

// setHealthCheck 对于已安装状态的应用实例，当应用支持健康检查时，开启健康检查，在其他状态下关闭健康检查
func (o *AppInstanceOperator) setHealthCheck(ctx context.Context, obj core.ApiObject) {
	appInstance := obj.(*v2.AppInstance)

	if appInstance.Status.Phase == core.PhaseInstalled {
		appObj, err := o.helper.V1.App.Get(context.TODO(), core.DefaultNamespace, appInstance.Spec.AppRef.Name)
		if err != nil {
			log.Error(err)
			return
		} else if appObj == nil {
			return
		}

		app := appObj.(*v1.App)
		for _, versionApp := range app.Spec.Versions {
			if versionApp.Version == appInstance.Spec.AppRef.Version {
				for _, supportAction := range versionApp.SupportActions {
					if supportAction == core.AppActionHealthcheck {
						o.enableHealthCheck(ctx, obj)
						return
					}
				}
			}
		}
	} else {
		o.disableHealthCheck(obj)
	}
}

// watchAndHandleJob 侦听任务的变更，并将任务的.Status.Phase变化交由handleJob处理，handleJob返回的bool值表示是否终止任务的侦听
func (o *AppInstanceOperator) watchAndHandleJob(ctx context.Context, jobName string, handleJob func(*v2.Job) bool) error {
	jobActionChan := o.helper.V2.Job.GetWatch(ctx, "", jobName)
	for jobAction := range jobActionChan {
		if jobAction.Obj == nil {
			err := e.Errorf("received nil object of job %s", jobName)
			log.Error(err)
			return err
		}
		switch jobAction.Type {
		case db.KVActionTypeDelete:
			err := e.Errorf("job %s has been deleted", jobName)
			log.Error(err)
			return err
		case db.KVActionTypeSet:
			if done := handleJob(jobAction.Obj.(*v2.Job)); done {
				return nil
			}
		}
	}
	return nil
}

// enableHealthCheck 开启健康检查
func (o *AppInstanceOperator) enableHealthCheck(ctx context.Context, obj core.ApiObject) {
	appInstance := obj.(*v2.AppInstance)

	healthcheckItem, ok := o.healthchecks.Get(appInstance.Metadata.Uid)

	// 如果健康检查已经存在，只更新检查参数
	if ok {
		healthcheckItem.InitialDelaySeconds = appInstance.Spec.LivenessProbe.InitialDelaySeconds
		healthcheckItem.PeriodSeconds = appInstance.Spec.LivenessProbe.PeriodSeconds
		healthcheckItem.TimeoutSeconds = appInstance.Spec.LivenessProbe.TimeoutSeconds
		o.healthchecks.Set(appInstance.Metadata.Uid, healthcheckItem)
		return
	}

	// 创建新的健康检查
	healthCheckCtx, healthCheckCancel := context.WithCancel(ctx)
	if o.healthchecks.Set(appInstance.Metadata.Uid, HealthCheckItem{
		Cancel: healthCheckCancel,
		LivenessProbe: v2.LivenessProbe{
			InitialDelaySeconds: appInstance.Spec.LivenessProbe.InitialDelaySeconds,
			PeriodSeconds:       appInstance.Spec.LivenessProbe.PeriodSeconds,
			TimeoutSeconds:      appInstance.Spec.LivenessProbe.TimeoutSeconds,
		},
	}) {
		return
	}

	go o.healthCheck(healthCheckCtx, appInstance)
}

// disableHealthCheck 关闭健康检查
func (o *AppInstanceOperator) disableHealthCheck(obj core.ApiObject) {
	appInstance := obj.(*v2.AppInstance)

	healthCheckItem, ok := o.healthchecks.Get(appInstance.Metadata.Uid)

	// 如果健康检查已经存在，则取消健康检查，并移除健康状态
	if ok {
		log.Debugf("cancel health check for app instance %s/%s", appInstance.Metadata.Namespace, appInstance.Metadata.Name)
		healthCheckItem.Cancel()
		o.healthchecks.Unset(appInstance.Metadata.Uid)
		appInstance.Status.UnsetCondition(core.ConditionTypeHealthy)
	}
}

// healthCheck 执行健康检查
func (o *AppInstanceOperator) healthCheck(ctx context.Context, obj core.ApiObject) {
	appInstance := obj.(*v2.AppInstance)

	// 退出时终止健康检查
	defer o.healthchecks.Unset(appInstance.Metadata.Uid)

	// 首次健康检查延迟
	time.Sleep(time.Duration(appInstance.Spec.LivenessProbe.InitialDelaySeconds) * time.Second)
	for {
		// 从内存中获取健康检查参数，因为应用实例的健康检查参数值可能会发生更新
		healthCheckItem, ok := o.healthchecks.Get(appInstance.Metadata.Uid)
		if !ok {
			return
		}
		appInstance.Spec.LivenessProbe = healthCheckItem.LivenessProbe

		select {
		case <-ctx.Done():
			return
		default:
			// 创建健康检查任务
			jobObj, err := o.setupJob(appInstance, core.EventActionHealthCheck)
			if err != nil {
				log.Error(err)
				o.failback(appInstance, core.EventActionHealthCheck, err.Error(), nil)
				time.Sleep(time.Duration(appInstance.Spec.LivenessProbe.PeriodSeconds) * time.Second)
				continue
			}
			job := jobObj.(*v2.Job)

			log.Debugf("running healthcheck of %s", appInstance.GetKey())

			// 侦听健康检查任务，出于性能方面考虑，此处不使用goroutine异步执行，即下一次健康检查的间隔计时是在上一次健康检查结束后才开始
			if err := o.watchAndHandleJob(ctx, job.Metadata.Name, func(job *v2.Job) bool {
				switch job.Status.Phase {
				case core.PhaseWaiting, core.PhaseRunning:
					// 任务运行中，不做任何处理
					return false
				case core.PhaseCompleted:
					log.Debugf("healthcheck succeed of %s", appInstance.GetKey())

					// 清除健康检查历史事件日志
					eventObjs, err := o.helper.V1.Event.List(context.TODO(), "")
					if err != nil {
						log.Error(err)
					} else {
						for _, eventObj := range eventObjs {
							event := eventObj.(*v1.Event)
							if event.Spec.ResourceRef.Namespace == appInstance.Metadata.Namespace && event.Spec.ResourceRef.Name == appInstance.Metadata.Name && event.Spec.ResourceRef.Kind == core.KindAppInstance && event.Spec.Action == core.EventActionHealthCheck {
								if _, err := o.helper.V1.Event.Delete(context.TODO(), "", event.Metadata.Name); err != nil {
									log.Error(err)
								}
							}
						}
					}

					// 如果原来处于非健康状态, 则更新为健康状态
					if appInstance.Status.GetCondition(core.ConditionTypeHealthy) != core.ConditionStatusTrue {
						appInstance.Status.SetCondition(core.ConditionTypeHealthy, core.ConditionStatusTrue)
						if _, err := o.helper.V2.AppInstance.UpdateStatus(appInstance.Metadata.Namespace, appInstance.Metadata.Name, appInstance.Status); err != nil {
							log.Error(err)
						}
					}

					// 记录事件完成
					if err := o.recordEvent(Event{
						BaseApiObj: appInstance.BaseApiObj,
						Action:     core.EventActionHealthCheck,
						Msg:        "",
						JobRef:     job.Metadata.Name,
						Phase:      core.PhaseCompleted,
					}); err != nil {
						log.Error(err)
					}
					return true
				case core.PhaseFailed:
					log.Warnf("healthcheck failed of %s", appInstance.GetKey())
					// 如果任务执行失败，将应用实例置为非健康状态
					o.failback(appInstance, core.EventActionHealthCheck, "", job)
					return true
				default:
					log.Warnf("unknown status phase '%s' of job '%s'", job.Status.Phase, job.GetKey())
					return false
				}
			}); err != nil {
				log.Error(err)
				o.failback(appInstance, core.EventActionHealthCheck, err.Error(), job)
			}
		}

		// 健康检查间隔
		time.Sleep(time.Duration(healthCheckItem.PeriodSeconds) * time.Second)
	}
}

// failback 操作失败回退
func (o AppInstanceOperator) failback(obj core.ApiObject, action string, reason string, job *v2.Job) {
	appInstance := obj.(*v2.AppInstance)

	var jobRef string
	if job != nil {
		jobRef = job.Metadata.Name
		if reason == "" {
			reason = job.Status.GetCondition(core.ConditionTypeRun)
		}
	}

	switch action {
	case core.EventActionInstall:
		if job == nil {
			// 关联任务为空时，恢复回未安装状态
			appInstance.SetStatusPhase(core.PhaseUninstalled)
		} else {
			appInstance.Status.SetCondition(core.ConditionTypeInstalled, reason)
			appInstance.Status.UnsetCondition(core.ConditionTypeHealthy)
			appInstance.SetStatusPhase(core.PhaseFailed)
		}
	case core.EventActionUninstall:
		if job == nil {
			// 关联任务为空时，恢复回已安装状态
			appInstance.SetStatusPhase(core.PhaseInstalled)
		} else {
			appInstance.Status.SetCondition(core.ConditionTypeInstalled, reason)
			appInstance.Status.UnsetCondition(core.ConditionTypeHealthy)
			appInstance.SetStatusPhase(core.PhaseFailed)
		}
	case core.EventActionConfigure:
		appInstance.Status.UnsetCondition(core.ConditionTypeHealthy)
		appInstance.Status.SetCondition(core.ConditionTypeConfigured, reason)
		appInstance.SetStatusPhase(core.PhaseInstalled)
	case core.EventActionUpgrade, core.EventActionRevert:
		if job == nil {
			// 关联任务为空时，可以直接恢复回原本状态
			appInstance.SetStatusPhase(core.PhaseInstalled)
		} else {
			appInstance.Status.UnsetCondition(core.ConditionTypeHealthy)
			appInstance.Status.SetCondition(core.ConditionTypeInstalled, reason)
			appInstance.SetStatusPhase(core.PhaseFailed)
		}
	case core.EventActionHealthCheck:
		// 清除健康检查历史事件日志
		eventObjs, err := o.helper.V1.Event.List(context.TODO(), "")
		if err != nil {
			log.Error(err)
		} else {
			for _, eventObj := range eventObjs {
				event := eventObj.(*v1.Event)
				if event.Spec.ResourceRef.Namespace == appInstance.Metadata.Namespace && event.Spec.ResourceRef.Name == appInstance.Metadata.Name && event.Spec.ResourceRef.Kind == core.KindAppInstance && event.Spec.Action == core.EventActionHealthCheck {
					if _, err := o.helper.V1.Event.Delete(context.TODO(), "", event.Metadata.Name); err != nil {
						log.Error(err)
					}
				}
			}
		}

		// 记录失败事件
		if err := o.recordEvent(Event{
			BaseApiObj: appInstance.BaseApiObj,
			Action:     action,
			Msg:        reason,
			JobRef:     jobRef,
			Phase:      core.PhaseFailed,
		}); err != nil {
			log.Error(err)
		}

		// 在健康状态发生变化时更新
		if appInstance.Status.GetCondition(core.ConditionTypeHealthy) != reason {
			appInstance.Status.SetCondition(core.ConditionTypeHealthy, reason)
			if _, err := o.helper.V2.AppInstance.Update(context.TODO(), appInstance, core.WithAllFields()); err != nil {
				log.Error(err)
			}
		}
		return
	}

	// 更新应用实例状态
	if _, err := o.helper.V2.AppInstance.UpdateStatus(appInstance.Metadata.Namespace, appInstance.Metadata.Name, appInstance.Status); err != nil {
		log.Error(err)
	}

	// 记录失败事件
	if err := o.recordEvent(Event{
		BaseApiObj: appInstance.BaseApiObj,
		Action:     action,
		Msg:        reason,
		JobRef:     jobRef,
		Phase:      core.PhaseFailed,
	}); err != nil {
		log.Error(err)
	}
}

// setupJob 根据操作行为构建任务
func (o *AppInstanceOperator) setupJob(obj core.ApiObject, action string) (core.ApiObject, error) {
	appInstance := obj.(*v2.AppInstance)

	// 获取应用实例对应的应用
	appObj, err := o.helper.V1.App.Get(context.TODO(), core.DefaultNamespace, appInstance.Spec.AppRef.Name)
	if err != nil {
		err = e.Errorf("failed to get referred app %s: %s", appInstance.Spec.AppRef.Name, err)
		log.Error(err)
		return nil, err
	} else if obj == nil {
		err := e.Errorf("referred app %s not found", appInstance.Spec.AppRef.Name)
		log.Error(err)
		return nil, err
	}
	app := appObj.(*v1.App)

	// 检索应用实例对应的应用版本
	var versionApp *v1.AppVersion
	for _, version := range app.Spec.Versions {
		if version.Version == appInstance.Spec.AppRef.Version {
			if !version.Enabled && action != core.EventActionHealthCheck && action != core.EventActionUninstall {
				err := e.Errorf("应用版本 %s 被禁用, 请检查部署包 %s 是否存在", version.Version, version.PkgRef)
				log.Error(err)
				return nil, err
			}
			versionApp = &version
			break
		}
	}
	if versionApp == nil {
		err := e.Errorf("referred app %s does not contain version of %s", appInstance.Spec.AppRef.Name, appInstance.Spec.AppRef.Version)
		log.Error(err)
		return nil, err
	}

	// 生成额外参数
	extraGlobalVars := make(map[string]interface{})
	switch app.Spec.Category {
	case core.AppCategoryCustomize:
		// 填充普通应用和算法插件额外参数
		package_dir, _ := filepath.Abs(filepath.Join(setting.PackageSetting.PkgPath, versionApp.PkgRef))
		extraGlobalVars["package_dir"] = package_dir
	}
	extraGlobalVars["app_name"] = app.Metadata.Name
	extraGlobalVars["deployer_data_dir"], _ = filepath.Abs(setting.AppSetting.DataDir)
	extraGlobalVars["app_instance_id"] = appInstance.Metadata.Uid

	// 构建inventory, playbook与配置文件
	plays := []v2.JobAnsiblePlay{}

	// 生成公共inventory内容
	commonInventoryStr, err := ansible.RenderCommonInventory()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	switch versionApp.Platform {
	case core.AppPlatformBareMetal:
		// 对于部署于裸机平台上的应用实例，以每个子模块为名构建group，子模块参数作为组参数，并将全局参数追加到每个组的参数
		for _, module := range appInstance.Spec.Modules {
			for replicaIndex, replicas := range module.Replicas {
				moduleAction := ModuleAction{
					ModuleName:    module.Name,
					AppVersion:    module.AppVersion,
					ReplicaIndex:  replicaIndex,
					HostActionMap: make(map[string]string),
				}

				for _, hostRef := range replicas.HostRefs {
					moduleAction.HostActionMap[hostRef] = strings.ToLower(action)
				}

				play, err := o.setupBareMetalJobPlay(appInstance, moduleAction, commonInventoryStr, extraGlobalVars, *app)
				if err != nil {
					log.Error(err)
					return nil, err
				}

				plays = append(plays, play)
			}
		}
	case core.AppPlatformK8s:
		// 对于部署于k8s平台上的应用实例，所有的模块都使用[k8s-master]

		// 获取k8s master节点
		host, err := o.helper.V1.K8sConfig.GetFirstMasterHost(appInstance.Spec.K8sRef)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		for _, module := range appInstance.Spec.Modules {
			for replicaIndex := range module.Replicas {
				play, err := o.setupK8sJobPlay(appInstance, module.Name, replicaIndex, action, commonInventoryStr, extraGlobalVars, *app, host)
				if err != nil {
					log.Error(err)
					return nil, err
				}

				plays = append(plays, play)
			}
		}
	}

	// 创建任务
	job := v2.NewJob()
	job.Metadata.Name = fmt.Sprintf("%s-%s-%s-%d", core.KindAppInstance, appInstance.Metadata.Name, action, time.Now().Unix())
	job.Spec.Exec.Type = core.JobExecTypeAnsible
	job.Spec.Exec.Ansible.Bin = setting.AnsibleSetting.Bin
	job.Spec.Exec.Ansible.Plays = plays
	if action == core.EventActionHealthCheck {
		job.Spec.Exec.Ansible.RecklessMode = true
	}
	if action == core.EventActionHealthCheck && appInstance.Spec.LivenessProbe.TimeoutSeconds > 0 {
		job.Spec.TimeoutSeconds = time.Duration(appInstance.Spec.LivenessProbe.TimeoutSeconds)
	} else {
		job.Spec.TimeoutSeconds = 3600
	}
	job.Spec.FailureThreshold = 1
	if _, err := o.helper.V2.Job.Create(context.TODO(), job); err != nil {
		log.Error(err)
		return nil, err
	}

	return job, nil
}

// setupUpgradeJob 构建应用实例的应用版本升级任务，表示将oldAppInstance升级/回退到newAppInstance
func (o AppInstanceOperator) setupUpgradeJob(oldAppInstance *v2.AppInstance, newAppInstance *v2.AppInstance) (*v2.Job, error) {
	// 获取应用
	appObj, err := o.helper.V1.App.Get(context.TODO(), core.DefaultNamespace, newAppInstance.Spec.AppRef.Name)
	if err != nil {
		log.Error(err)
		return nil, err
	} else if appObj == nil {
		err := e.Errorf("app %s not found", newAppInstance.Spec.AppRef)
		log.Error(err)
		return nil, err
	}
	app := appObj.(*v1.App)

	// 生成额外参数
	extraGlobalVars := make(map[string]interface{})
	extraGlobalVars["app_name"] = app.Metadata.Name
	extraGlobalVars["deployer_data_dir"], _ = filepath.Abs(setting.AppSetting.DataDir)
	extraGlobalVars["app_instance_id"] = newAppInstance.Metadata.Uid

	// 构建inventory, playbook与配置文件
	plays := []v2.JobAnsiblePlay{}

	// 生成公共inventory内容
	commonInventoryStr, err := ansible.RenderCommonInventory()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// 生成新旧版本间应用升级各模块所需执行的操作，主要分两个部分，分别是对旧版本的卸载和对新版本的安装配置
	oldModuleActions := []ModuleAction{}
	for _, oldModule := range oldAppInstance.Spec.Modules {
		newModule, ok := newAppInstance.GetModule(oldModule.Name)
		if !ok || (ok && oldModule.AppVersion != newModule.AppVersion) {
			// 当模块在新版本中不存在，或者存在但版本不同时，卸载旧模块中的所有节点
			for replicaIndex, replica := range oldModule.Replicas {
				hostActionMap := make(map[string]string)
				for _, hostRef := range replica.HostRefs {
					hostActionMap[hostRef] = core.AppActionUninstall
				}
				// 没有节点需要卸载则跳过
				if len(hostActionMap) <= 0 {
					continue
				}
				oldModuleActions = append(oldModuleActions, ModuleAction{
					ModuleName:    oldModule.Name,
					AppVersion:    oldModule.AppVersion,
					ReplicaIndex:  replicaIndex,
					HostActionMap: hostActionMap,
				})
			}
		} else {
			// 当旧版本模块仍然存在时，卸载被移除的节点
			for replicaIndex, replica := range oldModule.Replicas {
				hostActionMap := make(map[string]string)
				for _, hostRef := range replica.HostRefs {
					if !in(hostRef, newModule.Replicas[replicaIndex].HostRefs) {
						hostActionMap[hostRef] = core.AppActionUninstall
					}
				}
				// 没有节点需要卸载则跳过
				if len(hostActionMap) <= 0 {
					continue
				}
				oldModuleActions = append(oldModuleActions, ModuleAction{
					ModuleName:    oldModule.Name,
					AppVersion:    oldModule.AppVersion,
					ReplicaIndex:  replicaIndex,
					HostActionMap: hostActionMap,
				})
			}
		}
	}
	newModuleActions := []ModuleAction{}
	for _, newModule := range newAppInstance.Spec.Modules {
		// 检查新版本模块是否可用
		if !app.VersionEnabled(newModule.AppVersion) {
			return nil, e.Errorf("应用版本 %s 被禁用, 请检查部署包是否存在", newModule.AppVersion)
		}
		oldModule, ok := oldAppInstance.GetModule(newModule.Name)
		if !ok || (ok && newModule.AppVersion != oldModule.AppVersion) {
			// 当新模块在旧版本中不存在或者存在但版本不同，安装模块中的所有节点
			for replicaIndex, replica := range newModule.Replicas {
				hostActionMap := make(map[string]string)
				for _, hostRef := range replica.HostRefs {
					hostActionMap[hostRef] = core.AppActionInstall
				}
				// 没有节点需要安装则跳过
				if len(hostActionMap) <= 0 {
					continue
				}
				newModuleActions = append(newModuleActions, ModuleAction{
					ModuleName:    newModule.Name,
					AppVersion:    newModule.AppVersion,
					ReplicaIndex:  replicaIndex,
					HostActionMap: hostActionMap,
				})
			}
		} else {
			// 当新模块在旧版本中存在且版本相同，安装新加的节点，配置已存在的节点
			for replicaIndex, replica := range newModule.Replicas {
				hostActionMap := make(map[string]string)
				for _, hostRef := range replica.HostRefs {
					if !in(hostRef, oldModule.Replicas[replicaIndex].HostRefs) {
						hostActionMap[hostRef] = core.AppActionInstall
					} else {
						hostActionMap[hostRef] = core.AppActionConfigure
					}
				}
				// 没有节点需要配置安装则跳过
				if len(hostActionMap) <= 0 {
					continue
				}
				newModuleActions = append(newModuleActions, ModuleAction{
					ModuleName:    newModule.Name,
					AppVersion:    newModule.AppVersion,
					ReplicaIndex:  replicaIndex,
					HostActionMap: hostActionMap,
				})
			}
		}
	}

	// 填充任务Play
	switch app.Spec.Platform {
	case core.AppPlatformBareMetal:
		// 生成旧版本卸载Play
		for _, oldModuleAction := range oldModuleActions {
			play, err := o.setupBareMetalJobPlay(oldAppInstance, oldModuleAction, commonInventoryStr, extraGlobalVars, *app)
			if err != nil {
				log.Error(err)
				return nil, err
			}

			plays = append(plays, play)
		}

		// 生成新版本安装Play
		for _, newModuleAction := range newModuleActions {
			play, err := o.setupBareMetalJobPlay(newAppInstance, newModuleAction, commonInventoryStr, extraGlobalVars, *app)
			if err != nil {
				log.Error(err)
				return nil, err
			}

			plays = append(plays, play)
		}
	case core.AppPlatformK8s:
		// 对于部署于k8s平台上的应用实例，所有的模块都使用[k8s-master]

		// 获取k8s集群主节点
		masterHost, err := o.helper.V1.K8sConfig.GetFirstMasterHost(oldAppInstance.Spec.K8sRef)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		// 生成旧版本卸载Play
		for _, oldModuleAction := range oldModuleActions {
			play, err := o.setupK8sJobPlay(oldAppInstance, oldModuleAction.ModuleName, oldModuleAction.ReplicaIndex, core.EventActionUninstall, commonInventoryStr, extraGlobalVars, *app, masterHost)
			if err != nil {
				log.Error(err)
				return nil, err
			}

			plays = append(plays, play)
		}

		// 生成新版本安装Play
		for _, newModuleAction := range newModuleActions {
			play, err := o.setupK8sJobPlay(newAppInstance, newModuleAction.ModuleName, newModuleAction.ReplicaIndex, core.EventActionInstall, commonInventoryStr, extraGlobalVars, *app, masterHost)
			if err != nil {
				log.Error(err)
				return nil, err
			}

			plays = append(plays, play)
		}
	}

	// 创建任务
	job := v2.NewJob()
	job.Metadata.Name = fmt.Sprintf("%s-%s-%s-%s-to-%s-%d", core.KindAppInstance, newAppInstance.Metadata.Name, core.EventActionUpgrade, oldAppInstance.Spec.AppRef.Version, newAppInstance.Spec.AppRef.Version, time.Now().Unix())
	job.Spec.Exec.Type = core.JobExecTypeAnsible
	job.Spec.Exec.Ansible.Bin = setting.AnsibleSetting.Bin
	job.Spec.Exec.Ansible.Plays = plays
	job.Spec.TimeoutSeconds = 3600
	job.Spec.FailureThreshold = 1
	if _, err := o.helper.V2.Job.Create(context.TODO(), job); err != nil {
		log.Error(err)
		return nil, err
	}
	return job, nil
}

// setupBareMetalJobPlay 构建裸机任务play, 以模块切片为粒度构建
func (o AppInstanceOperator) setupBareMetalJobPlay(appInstance *v2.AppInstance, moduleAction ModuleAction, commonInventoryStr string, extraGlobalVars map[string]interface{}, app v1.App) (v2.JobAnsiblePlay, error) {
	var play v2.JobAnsiblePlay

	module, _ := appInstance.GetModule(moduleAction.ModuleName)
	if module.AppVersion == "" {
		module.AppVersion = appInstance.Spec.AppRef.Version
	}
	replica := module.Replicas[moduleAction.ReplicaIndex]

	// 校验每个主机所对应的模块版本是否可用
	versionApp, ok := app.GetVersion(module.AppVersion)
	if !ok {
		err := e.Errorf("version %s is not found in app %s", module.AppVersion, app.Metadata.Name)
		log.Error(err)
		return play, err
	} else if !versionApp.Enabled {
		for _, action := range moduleAction.HostActionMap {
			if action != core.AppActionHealthcheck && action != core.AppActionUninstall {
				err := e.Errorf("应用版本 %s 被禁用, 请检查部署包 %s 是否存在", versionApp.Version, versionApp.PkgRef)
				log.Error(err)
				return play, err
			}
		}
	}
	appModule, ok := versionApp.GetModule(module.Name)
	if !ok {
		err := e.Errorf("module %s is not found in app %s-%s", module.Name, app.Metadata.Name, module.AppVersion)
		log.Error(err)
		return play, err
	}

	inventory := make(map[string]ansible.InventoryGroup)
	groupVars := ansible.AppArgs{}

	// 填充全局主机别名
	hostAliases := make(map[string]string)
	for _, hostAlias := range appInstance.Spec.Global.HostAliases {
		hostAliases[hostAlias.Hostname] = hostAlias.IP
	}

	// 填充模块主机别名
	for _, hostAlias := range replica.HostAliases {
		hostAliases[hostAlias.Hostname] = hostAlias.IP
	}
	groupVars["host_aliases"] = hostAliases

	// 填充全局自定义参数
	for _, arg := range appInstance.Spec.Global.Args {
		groupVars.Set(arg.Name, arg.Value)
	}

	// 填充模块自定义参数
	for _, arg := range replica.Args {
		for _, appModule := range versionApp.Modules {
			if appModule.Name == module.Name {
				for _, appArg := range appModule.Args {
					if appArg.Name == arg.Name {
						// 根据参数类型填充参数值
						switch appArg.Type {
						case ArgTypeInteger:
							// 由于json反序列化会将数字类型统一转为float64, 因此需要先用类型推断转为float64再转为int
							var value interface{}
							switch v := arg.Value.(type) {
							case float64:
								switch appArg.Format {
								case ArgFormatInt32:
									value = int32(v)
								case ArgFormatInt64:
									value = int64(v)
								case ArgFormatPort:
									value = uint16(v)
								default:
									value = int64(v)
								}
							}
							groupVars.Set(arg.Name, value)
						case ArgTypeNumber:
							var value interface{}
							switch v := arg.Value.(type) {
							case float64:
								switch appArg.Format {
								case ArgFormatFloat:
									value = float32(v)
								case ArgFormatDouble:
									value = float64(v)
								default:
									value = float64(v)
								}
							}
							groupVars.Set(arg.Name, value)
						case ArgTypeBoolean:
							var value bool
							switch v := arg.Value.(type) {
							case bool:
								value = v
							}
							groupVars.Set(arg.Name, value)
						case ArgTypeString:
							var value interface{}
							switch v := arg.Value.(type) {
							case string:
								switch appArg.Format {
								case ArgFormatDate:
									value, _ = time.Parse(time.RFC3339, v)
									groupVars.Set(arg.Name, value)
								case ArgFormatArray:
									groupVars.Set(arg.Name, strings.Split(v, ";"))
								case ArgFormatGroupHost:
									hostRefs := strings.Split(v, ";")
									inventoryGroupHosts := make(map[string]ansible.InventoryHost)
									for _, hostRef := range hostRefs {
										hostObj, err := o.helper.V1.Host.Get(context.TODO(), "", hostRef)
										if err != nil {
											err = e.Errorf("failed to get host %s, %s", hostRef, err.Error())
											log.Error(err)
											return play, err
										} else if hostObj == nil {
											err := e.Errorf("host %s not found", hostRef)
											log.Error(err)
											return play, err
										}
										host := hostObj.(*v1.Host)
										inventoryGroupHosts[hostRef] = ansible.InventoryHost{
											"ansible_ssh_host": host.Spec.SSH.Host,
											"ansible_ssh_pass": host.Spec.SSH.Password,
											"ansible_ssh_port": host.Spec.SSH.Port,
											"ansible_ssh_user": host.Spec.SSH.User,
										}
									}
									inventory[arg.Name] = ansible.InventoryGroup{
										Hosts: inventoryGroupHosts,
										Vars:  make(map[string]interface{}),
									}
								default:
									groupVars.Set(arg.Name, v)
								}
							}
						default:
							groupVars.Set(arg.Name, arg.Value)
						}
						break
					}
				}
			}
		}
	}

	// 填充全局内置参数
	for varName, varValue := range extraGlobalVars {
		groupVars[varName] = varValue
	}

	// 填充模块内置参数
	for varName, varValue := range appModule.ExtraVars {
		groupVars[varName] = varValue
	}

	// 填充静态配置文件路径
	if replica.ConfigMapRef.Name != "" {
		configObj, err := o.helper.V1.ConfigMap.Get(context.TODO(), replica.ConfigMapRef.Namespace, replica.ConfigMapRef.Name)
		if err != nil {
			return play, err
		} else if configObj != nil {
			config := configObj.(*v1.ConfigMap)
			configFiles := []string{}
			for path := range config.Data {
				configFiles = append(configFiles, path)
			}
			groupVars["config_files"] = configFiles
		}
	}
	groupVars["configs_dir"] = ansible.ConfigsDir

	// 填充自定义配置文件路径
	if replica.AdditionalConfigs.ConfigMapRef.Name != "" {
		configObj, err := o.helper.V1.ConfigMap.Get(context.TODO(), replica.AdditionalConfigs.ConfigMapRef.Namespace, replica.AdditionalConfigs.ConfigMapRef.Name)
		if err != nil {
			return play, err
		} else if configObj != nil {
			config := configObj.(*v1.ConfigMap)
			configFiles := []string{}
			for path := range config.Data {
				configFiles = append(configFiles, path)
			}
			groupVars["additional_config_files"] = configFiles
		}
	}
	groupVars["additional_configs_dir"] = ansible.AdditionalConfigsDir

	// 关联配置文件
	configs := []v2.AnsibleConfig{}
	configs = append(configs, v2.AnsibleConfig{
		PathPrefix: ansible.ConfigsDir,
		ValueFrom: v2.ValueFrom{
			ConfigMapRef: replica.ConfigMapRef,
		},
	})
	configs = append(configs, v2.AnsibleConfig{
		PathPrefix: ansible.AdditionalConfigsDir,
		ValueFrom: v2.ValueFrom{
			ConfigMapRef: replica.AdditionalConfigs.ConfigMapRef,
		},
	})

	// 填充安装包路径
	groupVars["package_dir"], _ = filepath.Abs(filepath.Join(setting.PackageSetting.PkgPath, versionApp.PkgRef))

	// 填充模块名
	groupVars["module_name"] = moduleAction.ModuleName

	// 填充切片下标
	groupVars["replica_index"] = moduleAction.ReplicaIndex

	// 填充算法实例参数
	requestGPU := false
	supportGPUModels := []string{}
	if appModule.Resources.AlgorithmPlugin {
		var pluginName, pluginVersion, mediaType string
		for _, arg := range replica.Args {
			switch arg.Name {
			case "ALGORITHM_PLUGIN_NAME":
				switch v := arg.Value.(type) {
				case string:
					pluginName = v
				}
			case "ALGORITHM_PLUGIN_VERSION":
				switch v := arg.Value.(type) {
				case string:
					pluginVersion = v
				}
			case "ALGORITHM_MEDIA_TYPE":
				switch v := arg.Value.(type) {
				case string:
					mediaType = v
				}
			case "ALGORITHM_REQUEST_GPU":
				switch v := arg.Value.(type) {
				case bool:
					requestGPU = v
				}
			}
		}
		groupVars.Set("ALGORITHM_PLUGIN_NAME", pluginName)
		groupVars.Set("ALGORITHM_PLUGIN_VERSION", pluginVersion)
		groupVars.Set("ALGORITHM_MEDIA_TYPE", mediaType)

		// 填充算法插件参数
		pluginObj, err := o.helper.V1.App.Get(context.TODO(), core.DefaultNamespace, pluginName)
		if err != nil {
			log.Error(err)
			return play, e.Errorf("failed to get algorithm plugin %s", pluginName)
		}
		if pluginObj == nil {
			err := e.Errorf("algorithm plugin %s not found", pluginName)
			log.Error(err)
			return play, err
		}
		plugin := pluginObj.(*v1.App)
		for _, versionPlugin := range plugin.Spec.Versions {
			if versionPlugin.Version == pluginVersion {
				ap_package_dir, _ := filepath.Abs(filepath.Join(setting.PackageSetting.PkgPath, versionPlugin.PkgRef))
				groupVars["ap_package_dir"] = ap_package_dir

				for varKey, varValue := range versionPlugin.Modules[0].ExtraVars {
					groupVars[varKey] = varValue
				}
				supportGPUModels = versionPlugin.SupportGpuModels
			}
		}
	}

	inventoryGroupHosts := make(map[string]ansible.InventoryHost)
	tags := []string{}
	algorithmGPUIDs := make(map[string]interface{})

	// 构建inventory和tags
	for hostRef, action := range moduleAction.HostActionMap {
		hostObj, err := o.helper.V1.Host.Get(context.TODO(), "", hostRef)
		if err != nil {
			err := e.Errorf("failed to get referred host %s: %s", hostRef, err.Error())
			log.Error(err)
			return play, err
		} else if hostObj == nil {
			err := e.Errorf("host %s not found", hostRef)
			log.Error(err)
			return play, err
		}
		host := hostObj.(*v1.Host)

		// 构建group hosts
		inventoryGroupHosts[hostRef] = ansible.InventoryHost{
			"ansible_ssh_host": host.Spec.SSH.Host,
			"ansible_ssh_pass": host.Spec.SSH.Password,
			"ansible_ssh_port": host.Spec.SSH.Port,
			"ansible_ssh_user": host.Spec.SSH.User,
			// 额外的主机参数
			"NODE_NAME": host.Metadata.Name,
			"act":       action,
		}

		// 追加标签
		if !in(action, tags) {
			tags = append(tags, action)
		}

		// 在每台主机上寻找型号匹配且空闲的GPU与实例绑定
		if appModule.Resources.AlgorithmPlugin && requestGPU {
			gpuID, err := o.allocGPUSlot(appInstance, moduleAction.ModuleName, moduleAction.ReplicaIndex, host, supportGPUModels, action)
			if err != nil {
				log.Error(err)
				return play, err
			}
			algorithmGPUIDs[hostRef] = gpuID
		}
	}

	// 为算法实例设置gpu插槽序号
	groupVars["algorithm_gpu_ids"] = algorithmGPUIDs
	inventory[module.Name] = ansible.InventoryGroup{
		Hosts: inventoryGroupHosts,
	}

	// 当注册服务到k8s集群时，添加额外的[k8s-master]
	if appInstance.Spec.K8sRef != "" {
		k8sObj, err := o.helper.V1.K8sConfig.Get(context.TODO(), core.DefaultNamespace, appInstance.Spec.K8sRef)
		if err != nil {
			err = e.Errorf("failed to get k8s cluster %s, %s", appInstance.Spec.K8sRef, err.Error())
			log.Error(err)
			return play, err
		} else if k8sObj == nil {
			err := e.Errorf("k8s cluster '%s' not found", appInstance.Spec.K8sRef)
			log.Error(err)
			return play, err
		}
		k8s := k8sObj.(*v1.K8sConfig)

		k8sGroupHosts := make(map[string]ansible.InventoryHost)
		for _, k8sHost := range k8s.Spec.K8SMaster.Hosts {
			hostRef := k8sHost.ValueFrom.HostRef
			hostObj, err := o.helper.V1.Host.Get(context.TODO(), "", hostRef)
			if err != nil {
				err = e.Errorf("failed to get host %s, %s", hostRef, err.Error())
				log.Error(err)
				return play, err
			} else if hostObj == nil {
				err := e.Errorf("host %s not found", hostRef)
				log.Error(err)
				return play, err
			}
			host := hostObj.(*v1.Host)

			k8sGroupHosts[hostRef] = ansible.InventoryHost{
				"ansible_ssh_host": host.Spec.SSH.Host,
				"ansible_ssh_pass": host.Spec.SSH.Password,
				"ansible_ssh_user": host.Spec.SSH.User,
				"ansible_ssh_port": host.Spec.SSH.Port,
			}
		}

		inventory[ansible.ANSIBLE_GROUP_K8S_MASTER] = ansible.InventoryGroup{
			Hosts: k8sGroupHosts,
		}
	}

	// 生成group_vars与inventory
	groupVarsData, _ := yaml.Marshal(groupVars)
	inventoryStr, _ := ansible.RenderInventory(inventory)
	playbookStr, _ := ansible.RenderPlaybook([]ansible.Playbook{
		{
			Hosts:       []string{module.Name},
			Roles:       appModule.IncludeRoles,
			IncludeVars: []string{"group_vars.yml"},
		},
	})

	play.Name = fmt.Sprintf("%s-%d-%s", module.Name, moduleAction.ReplicaIndex, strings.Join(tags, ","))
	play.Configs = configs
	play.Tags = tags
	play.GroupVars = v2.AnsibleGroupVars{
		Value: string(groupVarsData),
	}
	play.Inventory = v2.AnsibleInventory{
		Value: inventoryStr + commonInventoryStr,
	}
	play.Playbook = v2.AnsiblePlaybook{
		Value: playbookStr,
	}
	return play, nil
}

// setupK8sJobPlay 构建k8s任务play，一个play对应着应用实例中的一个模块副本的一个action，并且应用实例的每个模块可以对应着不同的应用版本
func (o AppInstanceOperator) setupK8sJobPlay(appInstance *v2.AppInstance, moduleName string, replicaIndex int, action string, commonInventoryStr string, extraGlobalVars map[string]interface{}, app v1.App, host *v1.Host) (v2.JobAnsiblePlay, error) {
	var play v2.JobAnsiblePlay

	module, _ := appInstance.GetModule(moduleName)
	replica := module.Replicas[replicaIndex]

	versionApp, ok := app.GetVersion(module.AppVersion)
	if !ok {
		err := e.Errorf("version %s is not found in app %s", module.AppVersion, app.Metadata.Name)
		log.Error(err)
		return play, err
	} else if !versionApp.Enabled {
		err := e.Errorf("version %s is disabled in app %s", module.AppVersion, app.Metadata.Name)
		log.Error(err)
		return play, err
	}
	appModule, ok := versionApp.GetModule(module.Name)
	if !ok {
		err := e.Errorf("module %s is not found in app %s-%s", module.Name, app.Metadata.Name, module.AppVersion)
		log.Error(err)
		return play, err
	}

	inventory := map[string]ansible.InventoryGroup{
		"ansible.ANSIBLE_GROUP_K8S_MASTER": {
			Hosts: map[string]ansible.InventoryHost{
				host.Metadata.Name: {
					"ansible_ssh_host": host.Spec.SSH.Host,
					"ansible_ssh_pass": host.Spec.SSH.Password,
					"ansible_ssh_port": host.Spec.SSH.Port,
					"ansible_ssh_user": host.Spec.SSH.User,
				},
			},
		},
	}
	groupVars := ansible.AppArgs{}
	playbook := ansible.Playbook{}
	configs := []v2.AnsibleConfig{}
	playName := fmt.Sprintf("%s-%d", module.Name, replicaIndex)
	tags := []string{strings.ToLower(action)}

	for _, arg := range replica.Args {
		for _, appModule := range versionApp.Modules {
			if appModule.Name == module.Name {
				for _, appArg := range appModule.Args {
					if appArg.Name == arg.Name {
						// 根据参数类型填充参数值
						switch appArg.Type {
						case ArgTypeInteger:
							// 由于json反序列化会将数字类型统一转为float64, 因此需要先用类型推断转为float64再转为int
							var value interface{}
							switch v := arg.Value.(type) {
							case float64:
								switch appArg.Format {
								case ArgFormatInt32:
									value = int32(v)
								case ArgFormatInt64:
									value = int64(v)
								case ArgFormatPort:
									value = uint16(v)
								default:
									value = int64(v)
								}
							}
							groupVars.Set(arg.Name, value)
						case ArgTypeNumber:
							var value interface{}
							switch v := arg.Value.(type) {
							case float64:
								switch appArg.Format {
								case ArgFormatFloat:
									value = float32(v)
								case ArgFormatDouble:
									value = float64(v)
								default:
									value = float64(v)
								}
							}
							groupVars.Set(arg.Name, value)
						case ArgTypeBoolean:
							var value bool
							switch v := arg.Value.(type) {
							case bool:
								value = v
							}
							groupVars.Set(arg.Name, value)
						case ArgTypeString:
							var value interface{}
							switch v := arg.Value.(type) {
							case string:
								switch appArg.Format {
								case ArgFormatDate:
									value, _ = time.Parse(time.RFC3339, v)
									groupVars.Set(arg.Name, value)
								case ArgFormatArray:
									groupVars.Set(arg.Name, strings.Split(v, ";"))
								case ArgFormatGroupHost:
									hostRefs := strings.Split(v, ";")
									inventoryGroupHosts := make(map[string]ansible.InventoryHost)
									for _, hostRef := range hostRefs {
										hostObj, err := o.helper.V1.Host.Get(context.TODO(), "", hostRef)
										if err != nil {
											err = e.Errorf("failed to get host %s, %s", hostRef, err.Error())
											log.Error(err)
											return play, err
										} else if hostObj == nil {
											err := e.Errorf("host %s not found", hostRef)
											log.Error(err)
											return play, err
										}
										host := hostObj.(*v1.Host)
										inventoryGroupHosts[hostRef] = ansible.InventoryHost{
											"ansible_ssh_host": host.Spec.SSH.Host,
											"ansible_ssh_pass": host.Spec.SSH.Password,
											"ansible_ssh_port": host.Spec.SSH.Port,
											"ansible_ssh_user": host.Spec.SSH.User,
										}
									}
									inventory[arg.Name] = ansible.InventoryGroup{
										Hosts: inventoryGroupHosts,
										Vars:  make(map[string]interface{}),
									}
								}
							}
						default:
							groupVars.Set(arg.Name, arg.Value)
						}
						break
					}
				}
			}
		}
	}

	// 追加全局参数
	for _, arg := range appInstance.Spec.Global.Args {
		groupVars.Set(arg.Name, arg.Value)
	}

	// 追加额外组参数，都认定为内部参数
	for varName, varValue := range extraGlobalVars {
		groupVars[varName] = varValue
	}
	for varName, varValue := range appModule.ExtraVars {
		groupVars[varName] = varValue
	}
	playbook = ansible.Playbook{
		Hosts:       []string{ansible.ANSIBLE_GROUP_K8S_MASTER},
		Roles:       appModule.IncludeRoles,
		IncludeVars: []string{"group_vars.yml"},
	}

	configs = append(configs, v2.AnsibleConfig{
		PathPrefix: ansible.ConfigsDir,
		ValueFrom: v2.ValueFrom{
			ConfigMapRef: replica.ConfigMapRef,
		},
	})

	configs = append(configs, v2.AnsibleConfig{
		PathPrefix: ansible.AdditionalConfigsDir,
		ValueFrom: v2.ValueFrom{
			ConfigMapRef: replica.AdditionalConfigs.ConfigMapRef,
		},
	})

	groupVarsData, _ := yaml.Marshal(groupVars)
	inventoryStr, _ := ansible.RenderInventory(inventory)
	playbookStr, _ := ansible.RenderPlaybook([]ansible.Playbook{playbook})

	play.Name = playName
	play.Configs = configs
	play.Envs = []string{"act=" + strings.ToLower(action)}
	play.Tags = tags
	play.GroupVars = v2.AnsibleGroupVars{
		Value: string(groupVarsData),
	}
	play.Inventory = v2.AnsibleInventory{
		Value: inventoryStr + commonInventoryStr,
	}
	play.Playbook = v2.AnsiblePlaybook{
		Value: playbookStr,
	}
	return play, nil
}

// allocGPUSlot 将应用实例与主机上的GPU绑定并返回GPU插槽序号
func (o AppInstanceOperator) allocGPUSlot(appInstance *v2.AppInstance, moduleName string, replicaIndex int, host *v1.Host, supportGPUModels []string, action string) (int, error) {
	switch action {
	case core.AppActionInstall:
		// 绑定GPU
		for _, gpuInfo := range host.Spec.Info.GPUs {
			// 校验GPU型号
			modelMatched := false
			for _, model := range supportGPUModels {
				if model == gpuInfo.Type {
					modelMatched = true
					break
				}
			}
			if !modelMatched {
				continue
			}

			// 如果GPU未被使用，则绑定GPU
			gpuName := o.helper.V1.GPU.GetGPUName(host.Metadata.Name, gpuInfo.ID)
			gpuObj, err := o.helper.V1.GPU.Get(context.TODO(), "", gpuName)
			if err != nil {
				log.Error(err)
				return -1, err
			}
			if gpuObj == nil {
				err := e.Errorf("gpu %s not found", gpuName)
				log.Error(err)
				return -1, err
			}
			gpu := gpuObj.(*v1.GPU)
			if gpu.Status.Phase != core.PhaseBound {
				gpu.Spec.AppInstanceModuleRef = v1.AppInstanceModuleRef{
					AppInstanceRef: v1.AppInstanceRef{
						Namespace: appInstance.Metadata.Namespace,
						Name:      appInstance.Metadata.Name,
					},
					Module:  moduleName,
					Replica: replicaIndex,
				}
				gpu.Status.Phase = core.PhaseBound
				if _, err := o.helper.V1.GPU.Update(context.TODO(), gpu, core.WithAllFields()); err != nil {
					log.Error(err)
					return -1, err
				}
				return gpuInfo.ID, nil
			}
		}
	case core.AppActionUninstall, core.AppActionConfigure, core.AppActionHealthcheck:
		// 获取已绑定的GPU
		gpuObjs, err := o.helper.V1.GPU.List(context.TODO(), "")
		if err != nil {
			log.Error(err)
			return -1, err
		}
		for _, gpuObj := range gpuObjs {
			gpu := gpuObj.(*v1.GPU)
			if gpu.Spec.AppInstanceModuleRef.Namespace == appInstance.Metadata.Namespace && gpu.Spec.AppInstanceModuleRef.Name == appInstance.Metadata.Name && gpu.Spec.AppInstanceModuleRef.Module == moduleName && gpu.Spec.AppInstanceModuleRef.Replica == replicaIndex {
				return gpu.Spec.Info.ID, nil
			}
		}
	}
	return -1, e.Errorf("host %s is not bound with gpu types %v", host.Metadata.Name, supportGPUModels)
}

// uninstallAppInstance 卸载应用实例
func (o AppInstanceOperator) uninstallAppInstance(ctx context.Context, appInstance *v2.AppInstance) {
	action := core.EventActionUninstall

	// 创建卸载任务
	jobObj, err := o.setupJob(appInstance, action)
	if err != nil {
		log.Errorf("setup %s job failed of %s: %s", action, appInstance.GetKey(), err)
		o.failback(appInstance, action, err.Error(), nil)
		return
	}
	job := jobObj.(*v2.Job)

	// 记录卸载事件开始，由于事件记录非必要流程，因此事件记录失败不会中断执行过程
	if err := o.recordEvent(Event{
		BaseApiObj: appInstance.BaseApiObj,
		Action:     action,
		Msg:        "",
		JobRef:     job.Metadata.Name,
		Phase:      core.PhaseWaiting,
	}); err != nil {
		log.Error(err)
	}

	// 侦听卸载任务的状态，并在任务执行完成时将应用实例状态置为已卸载
	o.watchAndHandleJob(ctx, job.Metadata.Name, func(job *v2.Job) bool {
		switch job.Status.Phase {
		case core.PhaseCompleted:
			// 释放算法实例GPU
			if err := o.releaseGPU(appInstance); err != nil {
				log.Error(err)
				o.failback(appInstance, core.EventActionUninstall, err.Error(), job)
				return true
			}

			// 如果初始化任务执行成功, 将应用实例状态更新为已卸载并结束任务侦听
			appInstance.Status.SetCondition(core.ConditionTypeInstalled, core.ConditionStatusFalse)
			appInstance.Status.UnsetCondition(core.ConditionTypeHealthy)
			appInstance.SetStatusPhase(core.PhaseUninstalled)
			if _, err := o.helper.V2.AppInstance.Update(context.TODO(), appInstance, core.WithAllFields()); err != nil {
				log.Error(err)
				return true
			}

			// 记录事件完成
			if err := o.recordEvent(Event{
				BaseApiObj: appInstance.BaseApiObj,
				Action:     core.EventActionUninstall,
				Msg:        "",
				JobRef:     job.Metadata.Name,
				Phase:      core.PhaseCompleted,
			}); err != nil {
				log.Error(err)
			}
			return true
		case core.PhaseFailed:
			o.failback(appInstance, core.EventActionUninstall, "", job)
			return true
		case core.PhaseWaiting, core.PhaseRunning:
			// 处于运行中状态不做任何处理
			return false
		default:
			log.Warnf("unknown status phase '%s' of job '%s'", job.Status.Phase, job.Metadata.Name)
			return false
		}
	})
}

// installAppInstance 安装应用实例
func (o AppInstanceOperator) installAppInstance(ctx context.Context, appInstance *v2.AppInstance) {
	action := core.EventActionInstall

	// 创建安装任务
	jobObj, err := o.setupJob(appInstance, action)
	if err != nil {
		log.Errorf("setup %s job failed of %s: %s", action, appInstance.GetKey(), err)
		o.failback(appInstance, action, err.Error(), nil)
		return
	}
	job := jobObj.(*v2.Job)

	// 记录安装事件开始，由于事件记录非必要流程，因此事件记录失败不会中断执行过程
	if err := o.recordEvent(Event{
		BaseApiObj: appInstance.BaseApiObj,
		Action:     action,
		Msg:        "",
		JobRef:     job.Metadata.Name,
		Phase:      core.PhaseWaiting,
	}); err != nil {
		log.Error(err)
	}

	// 侦听安装任务的状态，并在任务执行完成时将应用实例状态置为已就绪
	o.watchAndHandleJob(ctx, job.Metadata.Name, func(job *v2.Job) bool {
		switch job.Status.Phase {
		case core.PhaseWaiting, core.PhaseRunning:
			// 任务运行中，不做任何处理
			return false
		case core.PhaseCompleted:
			// 如果任务执行成功, 将应用实例状态更新为已安装并结束任务侦听
			appInstance.Status.SetCondition(core.ConditionTypeInstalled, core.ConditionStatusTrue)
			appInstance.SetStatusPhase(core.PhaseInstalled)
			if _, err := o.helper.V2.AppInstance.Update(context.TODO(), appInstance, core.WithAllFields()); err != nil {
				log.Error(err)
				return true
			}

			// 记录事件完成
			if err := o.recordEvent(Event{
				BaseApiObj: appInstance.BaseApiObj,
				Action:     core.EventActionInstall,
				Msg:        "",
				JobRef:     job.Metadata.Name,
				Phase:      core.PhaseCompleted,
			}); err != nil {
				log.Error(err)
			}
			return true
		case core.PhaseFailed:
			o.failback(appInstance, core.EventActionInstall, "", job)
			return true
		default:
			log.Warnf("unknown status phase '%s' of job '%s'", job.Status.Phase, job.Metadata.Name)
			return false
		}
	})
}

// updateAppInstance 更新应用实例，更新动作包括：升级，回滚和配置
func (o AppInstanceOperator) updateAppInstance(ctx context.Context, obj core.ApiObject, eventAction string) {
	newAppInstance := obj.(*v2.AppInstance)

	// 获取升级前的结构
	oldObj, err := o.revisioner.GetLastRevision(ctx, newAppInstance.Metadata.Namespace, newAppInstance.Metadata.Name)
	if err != nil {
		log.Error(err)
		return
	} else if oldObj == nil {
		log.Errorf("no revision of %s exists", newAppInstance.GetKey())
		return
	}
	oldAppInstance := oldObj.(*v2.AppInstance)

	var eventMsg string
	if newAppInstance.Spec.AppRef.Version != oldAppInstance.Spec.AppRef.Version {
		eventMsg = fmt.Sprintf("从 %s 到 %s", oldAppInstance.Spec.AppRef.Version, newAppInstance.Spec.AppRef.Version)
	}

	job, err := o.setupUpgradeJob(oldAppInstance, newAppInstance)
	if err != nil {
		log.Error(err)
		o.failback(oldAppInstance, eventAction, err.Error(), job)
		return
	}

	// 记录事件开始
	if err := o.recordEvent(Event{
		BaseApiObj: newAppInstance.BaseApiObj,
		Action:     eventAction,
		Msg:        eventMsg,
		JobRef:     job.Metadata.Name,
		Phase:      core.PhaseWaiting,
	}); err != nil {
		log.Error(err)
	}

	// 监听升级job
	o.watchAndHandleJob(ctx, job.Metadata.Name, func(job *v2.Job) bool {
		switch job.Status.Phase {
		case core.PhaseWaiting, core.PhaseRunning:
			// 任务运行中，不做任何处理
			return false
		case core.PhaseCompleted:
			// 记录事件完成
			if err := o.recordEvent(Event{
				BaseApiObj: newAppInstance.BaseApiObj,
				Action:     eventAction,
				Msg:        eventMsg,
				JobRef:     job.Metadata.Name,
				Phase:      core.PhaseCompleted,
			}); err != nil {
				log.Error(err)
			}
			delete(newAppInstance.Metadata.Annotations, core.AnnotationPrefix+"upgrade/last-applied-configuration")

			// 如果任务执行成功, 将应用实例置为Installed状态
			newAppInstance.Status.SetCondition(core.ConditionTypeInstalled, core.ConditionStatusTrue)
			newAppInstance.SetStatusPhase(core.PhaseInstalled)

			if _, err := o.helper.V2.AppInstance.Update(context.TODO(), newAppInstance, core.WithAllFields()); err != nil {
				log.Error(err)
			}
			return true
		case core.PhaseFailed:
			o.failback(newAppInstance, eventAction, eventMsg, job)
			return true
		default:
			log.Warnf("unknown status phase '%s' of job '%s'", job.Status.Phase, job.GetKey())
			return false
		}
	})
}

// releaseGPU 释放应用实例中绑定的所有GPU
func (o AppInstanceOperator) releaseGPU(appInstance *v2.AppInstance) error {
	gpuObjs, err := o.helper.V1.GPU.List(context.TODO(), "")
	if err != nil {
		log.Error(err)
		return err
	}
	for _, gpuObj := range gpuObjs {
		gpu := gpuObj.(*v1.GPU)
		if gpu.Spec.AppInstanceModuleRef.Namespace == appInstance.Metadata.Namespace && gpu.Spec.AppInstanceModuleRef.Name == appInstance.Metadata.Name {
			gpu.Spec.AppInstanceModuleRef = v1.AppInstanceModuleRef{}
			gpu.Status.Phase = core.PhaseWaiting
			log.Debugf("%+v", gpu)
			if _, err := o.helper.V1.GPU.Update(context.TODO(), gpu, core.WithAllFields()); err != nil {
				log.Error(err)
				return err
			}
		}
	}
	return nil
}

// in 判断数组中是否存在目标项
func in(target string, array []string) bool {
	for _, item := range array {
		if target == item {
			return true
		}
	}
	return false
}

// NewAppInstanceOperator 创建应用实例管理器
func NewAppInstanceOperator() *AppInstanceOperator {
	o := &AppInstanceOperator{
		BaseOperator: NewBaseOperator(v2.NewAppInstanceRegistry()),
		revisioner:   v2.NewAppInstanceRevision(),
	}
	o.SetFinalizeFunc(o.finalizeAppInstance)
	o.SetHandleFunc(o.handleAppInstance)
	o.healthchecks = NewHealthChecks()
	return o
}
