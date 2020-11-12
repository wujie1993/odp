package v2

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"text/template"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

func (obj AppInstance) GetModule(moduleName string) (AppInstanceModule, bool) {
	for _, module := range obj.Spec.Modules {
		if module.Name == moduleName {
			return module, true
		}
	}
	return AppInstanceModule{}, false
}

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

func (obj AppInstance) GetGlobalArgValue(argName string) interface{} {
	for _, arg := range obj.Spec.Global.Args {
		if arg.Name == argName {
			return arg.Value
		}
	}
	return nil
}

func (obj AppInstance) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

func (obj *AppInstance) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

func (obj AppInstance) SpecHash() string {
	for moduleIndex, module := range obj.Spec.Modules {
		for replicaIndex := range module.Replicas {
			obj.Spec.Modules[moduleIndex].Replicas[replicaIndex].Notes = ""
		}
	}
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

type AppInstanceRegistry struct {
	registry.Registry
}

func appInstancePreCreate(obj core.ApiObject) error {
	appInstance := obj.(*AppInstance)
	appInstance.Metadata.Finalizers = []string{core.FinalizerCleanRefEvent, core.FinalizerReleaseRefGPU, core.FinalizerCleanRefConfigMap}
	return nil
}

func appInstancePostCreate(obj core.ApiObject) error {
	hostRegistry := v1.NewHostRegistry()

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
			host := hostObj.(*v1.Host)

			host.Spec.Plugins = append(host.Spec.Plugins, v1.HostPlugin{
				AppInstanceRef: v1.AppInstanceRef{
					Namespace: appInstance.Metadata.Namespace,
					Name:      appInstance.Metadata.Name,
				},
				AppRef: v1.AppRef{
					Name:    appInstance.Spec.AppRef.Name,
					Version: appInstance.Spec.AppRef.Version,
				},
			})

			if _, err := hostRegistry.Update(context.TODO(), host, core.WithStatus()); err != nil {
				log.Error(err)
				return err
			}
		}
	}

	return nil
}

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

	switch appInstance.Spec.Action {
	case core.AppActionConfigure:
		// 非Installed状态禁止配置操作
		if oldAppInstance.Status.Phase != core.PhaseInstalled {
			err := e.Errorf("not allow to configure when status phase not Installed")
			log.Error(err)
			return err
		}

		var versionApp v1.AppVersion
		for _, appVersion := range app.Spec.Versions {
			if appVersion.Version == appInstance.Spec.AppRef.Version {
				versionApp = appVersion
				break
			}
		}
		// 重置关联应用
		appInstance.Spec.AppRef = oldAppInstance.Spec.AppRef
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

func appInstancePostDelete(obj core.ApiObject) error {
	hostRegistry := v1.NewHostRegistry()

	appInstance := obj.(*AppInstance)

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
			host := hostObj.(*v1.Host)

			for index, plugin := range host.Spec.Plugins {
				if plugin.AppRef.Name == appInstance.Spec.AppRef.Name && plugin.AppRef.Version == appInstance.Spec.AppRef.Version {
					host.Spec.Plugins = append(host.Spec.Plugins[:index], host.Spec.Plugins[index+1:]...)
				}
			}

			if _, err := hostRegistry.Update(context.TODO(), host, core.WithStatus()); err != nil {
				log.Error(err)
				return err
			}
		}
	}

	return nil
}

func appInstanceValidate(obj core.ApiObject) error {
	hostRegistry := v1.NewHostRegistry()
	appRegistry := v1.NewAppRegistry()

	appInstance := obj.(*AppInstance)

	// 验证应用是否存在
	appObj, err := appRegistry.Get(context.TODO(), core.DefaultNamespace, appInstance.Spec.AppRef.Name)
	if err != nil {
		return err
	} else if appObj == nil {
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
	if !appVersionExist {
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

func appInstanceMutate(obj core.ApiObject) error {
	appInstance := obj.(*AppInstance)

	appInstanceRegistry := NewAppInstanceRegistry()
	// 获取更新前的应用实例
	oldObj, err := appInstanceRegistry.Get(context.TODO(), appInstance.Metadata.Namespace, appInstance.Metadata.Name)
	if err != nil {
		log.Error(err)
		return err
	} else if oldObj == nil {
		return nil
	}
	oldAppInstance := oldObj.(*AppInstance)

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

	// 填充应用实例分类
	appInstance.Spec.Category = app.Spec.Category

	// 填充健康检查配置项
	if appInstance.Spec.LivenessProbe.InitialDelaySeconds < 0 {
		appInstance.Spec.LivenessProbe.InitialDelaySeconds = 10
	}
	if appInstance.Spec.LivenessProbe.PeriodSeconds < 10 {
		appInstance.Spec.LivenessProbe.PeriodSeconds = 30
	}
	if appInstance.Spec.LivenessProbe.TimeoutSeconds < 10 {
		appInstance.Spec.LivenessProbe.TimeoutSeconds = 30
	}

	for moduleIndex, module := range appInstance.Spec.Modules {
		for replicaIndex := range module.Replicas {
			// 移除Notes
			appInstance.Spec.Modules[moduleIndex].Replicas[replicaIndex].Notes = ""
		}

		// 在安装或升级应用实例时，更新模块版本
		switch appInstance.Spec.Action {
		case core.AppActionInstall:
			appInstance.Spec.Modules[moduleIndex].AppVersion = appInstance.Spec.AppRef.Version
		case core.AppActionUpgrade:
			// 从旧版本模块中恢复模块版本
			for _, oldModule := range oldAppInstance.Spec.Modules {
				if module.Name == oldModule.Name {
					appInstance.Spec.Modules[moduleIndex].AppVersion = oldModule.AppVersion
				}
			}
			// 更新需要升级的模块版本
			for _, appModule := range versionApp.Modules {
				if module.Name == appModule.Name && !appModule.SkipUpgrade {
					appInstance.Spec.Modules[moduleIndex].AppVersion = appInstance.Spec.AppRef.Version
				}
			}
		}
	}

	return nil
}

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

func NewAppInstance() *AppInstance {
	appInstance := new(AppInstance)
	appInstance.Init(ApiVersion, core.KindAppInstance)
	appInstance.Spec.LivenessProbe.InitialDelaySeconds = 10
	appInstance.Spec.LivenessProbe.PeriodSeconds = 30
	appInstance.Spec.LivenessProbe.TimeoutSeconds = 30
	return appInstance
}

func NewAppInstanceRegistry() AppInstanceRegistry {
	r := AppInstanceRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindAppInstance), true),
	}
	r.SetValidateHook(appInstanceValidate)
	r.SetMutateHook(appInstanceMutate)
	r.SetDecorateHook(appInstanceDecorate)
	r.SetPreCreateHook(appInstancePreCreate)
	r.SetPostCreateHook(appInstancePostCreate)
	r.SetPreUpdateHook(appInstancePreUpdate)
	r.SetPostDeleteHook(appInstancePostDelete)
	return r
}
