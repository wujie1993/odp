package v2

import (
	"bytes"
	"context"
	"io/ioutil"
	"path"
	"text/template"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/util"
)

// AppInstanceRegistry 应用实例存储器
// +namespaced=true
type AppInstanceRegistry struct {
	registry.Registry
}

// appInstancePostCreate 自定义应用实例创建后逻辑
func appInstancePostCreate(obj core.ApiObject) error {
	hostRegistry := NewHostRegistry()

	appInstance := obj.(*AppInstance)

	if appInstance.Spec.Category == core.AppCategoryHostPlugin && len(appInstance.Spec.Modules) > 0 && len(appInstance.Spec.Modules[0].Replicas) > 0 {
		// 只针对 第一个模块取主机
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
			host := hostObj.(*Host)

			host.Info.Plugins = append(host.Info.Plugins, HostPlugin{
				AppInstanceRef: AppInstanceRef{
					Namespace: appInstance.Metadata.Namespace,
					Name:      appInstance.Metadata.Name,
				},
				AppRef: AppRef{
					Name:    appInstance.Spec.AppRef.Name,
					Version: appInstance.Spec.AppRef.Version,
				},
			})

			if _, err := hostRegistry.Update(context.TODO(), host, core.WithAllFields()); err != nil {
				log.Error(err)
				return err
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

	if appInstance.Spec.Action == core.AppActionUpdate {
		if appInstance.Spec.AppRef.Version == oldAppInstance.Spec.AppRef.Version {
			appInstance.Spec.Action = core.AppActionConfigure
		} else {
			appInstance.Spec.Action = core.AppActionUpgrade
		}
	}

	switch appInstance.Spec.Action {
	case core.AppActionConfigure:
		appRegistry := v1.NewAppRegistry()
		// 获取关联的应用
		appObj, err := appRegistry.Get(context.TODO(), core.DefaultNamespace, appInstance.Spec.AppRef.Name)
		if err != nil {
			log.Error(err)
			return err
		}
		if appObj == nil {
			return nil
		}
		app := appObj.(*v1.App)

		// 非Installed状态禁止操作
		if oldAppInstance.Status.Phase != core.PhaseInstalled {
			err := e.Errorf("not allow to configure when status phase not Installed")
			log.Error(err)
			return err
		}

		action := appInstance.Spec.Action
		appInstance.Spec.Action = ""
		if oldAppInstance.SpecHash() != appInstance.SpecHash() {
			appInstance.Spec.Action = action
		}

		var versionApp v1.AppVersion
		for _, appVersion := range app.Spec.Versions {
			if appVersion.Version == appInstance.Spec.AppRef.Version {
				versionApp = appVersion
				break
			}
		}
		// 重置禁止修改的模块参数
		for moduleIndex, module := range appInstance.Spec.Modules {
			for replicaIndex, replica := range module.Replicas {
				for argIndex, arg := range replica.Args {
					for _, appModule := range versionApp.Modules {
						if appModule.Name != module.Name {
							continue
						}
						for _, appArg := range appModule.Args {
							if appArg.Name != arg.Name {
								continue
							}
							if !appArg.Modifiable {
								appInstance.Spec.Modules[moduleIndex].Replicas[replicaIndex].Args[argIndex].Value = oldAppInstance.GetModuleReplicaArgValue(module.Name, replicaIndex, arg.Name)
							}
							if appArg.Readonly {
								appInstance.Spec.Modules[moduleIndex].Replicas[replicaIndex].Args[argIndex].Value = appArg.Default
							}
						}
					}
				}
			}
		}
		// 重置禁止修改的全局参数
		for argIndex, arg := range appInstance.Spec.Global.Args {
			for _, appArg := range versionApp.Global.Args {
				if arg.Name != appArg.Name {
					continue
				}
				if !appArg.Modifiable {
					appInstance.Spec.Global.Args[argIndex].Value = oldAppInstance.GetGlobalArgValue(arg.Name)
				}
				if appArg.Readonly {
					appInstance.Spec.Global.Args[argIndex].Value = appArg.Default
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
	hostRegistry := v1.NewHostRegistry()
	appRegistry := v1.NewAppRegistry()

	appInstance := obj.(*AppInstance)

	// 验证应用是否存在
	appObj, err := appRegistry.Get(context.TODO(), core.DefaultNamespace, appInstance.Spec.AppRef.Name)
	if err != nil {
		return err
	} else if appObj == nil && appInstance.Status.Phase != core.PhaseDeleting {
		return e.Errorf("referred app %s not found", appInstance.Spec.AppRef.Name)
	}
	app := appObj.(*v1.App)

	// 验证应用版本是否存在
	var appVersionExist bool
	for _, versionApp := range app.Spec.Versions {
		if versionApp.Version == appInstance.Spec.AppRef.Version {
			appVersionExist = true
			break
		}
	}
	if !appVersionExist && appInstance.Status.Phase != core.PhaseDeleting {
		return e.Errorf("referred app version %s not found", appInstance.Spec.AppRef.Version)
	}

	switch app.Spec.Category {
	case core.AppCategoryHostPlugin:
		for _, module := range appInstance.Spec.Modules {
			if len(module.Replicas) > 1 {
				err := e.Errorf("host plugin is not allow to have more than 1 replicas in modules")
				log.Error(err)
				return err
			}
			// 限制主机插件在一台主机上只能安装一个
			for _, hostRef := range module.Replicas[0].HostRefs {
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
				host := hostObj.(*v1.Host)

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
		return e.Errorf("algorithm plugin is not allow to create app instance")
	}

	return nil
}

// appInstanceMutate 自定义应用实例内容写入填充逻辑
func appInstanceMutate(obj core.ApiObject) error {
	appInstance := obj.(*AppInstance)

	appRegistry := v1.NewAppRegistry()
	// 获取关联的应用
	appObj, err := appRegistry.Get(context.TODO(), core.DefaultNamespace, appInstance.Spec.AppRef.Name)
	if err != nil {
		log.Error(err)
		return err
	}
	if appObj != nil {
		app := appObj.(*v1.App)

		// 填充应用实例分类
		appInstance.Spec.Category = app.Spec.Category
	}

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

	cmRegistry := v1.NewConfigMapRegistry()
	for moduleIndex, module := range appInstance.Spec.Modules {
		for replicaIndex, replica := range module.Replicas {
			// 无需存储Notes字段
			appInstance.Spec.Modules[moduleIndex].Replicas[replicaIndex].Notes = ""

			// 填充配置文件哈希值，确保配置文件更新时，应用实例内容也发生更新
			if replica.ConfigMapRef.Name != "" {
				cm, err := cmRegistry.Get(context.TODO(), replica.ConfigMapRef.Namespace, replica.ConfigMapRef.Name)
				if err != nil {
					log.Error(err)
					return err
				}

				if cm != nil {
					appInstance.Spec.Modules[moduleIndex].Replicas[replicaIndex].ConfigMapRef.Hash = cm.SpecHash()
					appInstance.Spec.Modules[moduleIndex].Replicas[replicaIndex].ConfigMapRef.Revision = cm.GetMetadata().ResourceVersion
				}
			} else {
				appInstance.Spec.Modules[moduleIndex].Replicas[replicaIndex].ConfigMapRef = ConfigMapRef{}
			}
			if replica.AdditionalConfigs.Enabled && replica.AdditionalConfigs.ConfigMapRef.Name != "" {
				cm, err := cmRegistry.Get(context.TODO(), replica.AdditionalConfigs.ConfigMapRef.Namespace, replica.AdditionalConfigs.ConfigMapRef.Name)
				if err != nil {
					log.Error(err)
					return err
				}

				if cm != nil {
					appInstance.Spec.Modules[moduleIndex].Replicas[replicaIndex].AdditionalConfigs.ConfigMapRef.Hash = cm.SpecHash()
				}
			} else {
				appInstance.Spec.Modules[moduleIndex].Replicas[replicaIndex].AdditionalConfigs = AdditionalConfigs{}
			}
			if replica.Args == nil {
				appInstance.Spec.Modules[moduleIndex].Replicas[replicaIndex].Args = []AppInstanceArgs{}
			}
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
		appRegistry := v1.NewAppRegistry()

		// 获取关联的应用
		appObj, err := appRegistry.Get(context.TODO(), core.DefaultNamespace, appInstance.Spec.AppRef.Name)
		if err != nil {
			log.Error(err)
			return err
		}
		if appObj == nil {
			return nil
		}
		app := appObj.(*v1.App)

		var versionApp v1.AppVersion
		for _, version := range app.Spec.Versions {
			if version.Version == appInstance.Spec.AppRef.Version {
				versionApp = version
			}
		}

		hostRegistry := v1.NewHostRegistry()
		for moduleIndex, module := range appInstance.Spec.Modules {
			for _, appModule := range versionApp.Modules {
				if module.Name != appModule.Name {
					continue
				}
				for replicaIndex, replica := range module.Replicas {
					notes := make(map[string]interface{})
					// 填充Notes
					if versionApp.Platform == core.AppPlatformBareMetal {
						hosts := []string{}
						for _, hostRef := range replica.HostRefs {
							hostObj, err := hostRegistry.Get(context.TODO(), "", hostRef)
							if err != nil {
								log.Error(err)
								return err
							} else if hostObj == nil {
								continue
							}
							host := hostObj.(*v1.Host)
							hosts = append(hosts, host.Spec.SSH.Host)
						}
						notes["Hosts"] = hosts
					}
					args := make(map[string]interface{})
					for _, arg := range replica.Args {
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

						appInstance.Spec.Modules[moduleIndex].Replicas[replicaIndex].Notes = buffer.String()
					}
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
		core.FinalizerCleanRevision,
		core.FinalizerCleanHostPlugin,
	})
	r.SetValidateHook(appInstanceValidate)
	r.SetMutateHook(appInstanceMutate)
	r.SetDecorateHook(appInstanceDecorate)
	r.SetPostCreateHook(appInstancePostCreate)
	r.SetPreUpdateHook(appInstancePreUpdate)
	r.SetRevisioner(NewAppInstanceRevision())
	return r
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
