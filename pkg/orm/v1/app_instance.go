package v1

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
)

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

func (obj AppInstance) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

func (obj *AppInstance) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

func (obj AppInstance) SpecHash() string {
	for moduleIndex := range obj.Spec.Modules {
		obj.Spec.Modules[moduleIndex].Notes = ""
	}
	obj.Spec.LivenessProbe = LivenessProbe{}
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// +namespaced=true
type AppInstanceRegistry struct {
	registry.Registry
}

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

			if _, err := hostRegistry.Update(context.TODO(), host, core.WithStatus()); err != nil {
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

			if _, err := hostRegistry.Update(context.TODO(), host, core.WithStatus()); err != nil {
				log.Error(err)
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

func appInstancePostDelete(obj core.ApiObject) error {
	hostRegistry := NewHostRegistry()

	appInstance := obj.(*AppInstance)

	if appInstance.Spec.Category == core.AppCategoryHostPlugin && len(appInstance.Spec.Modules) > 0 {
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
	} else if appInstance.Spec.Category == core.AppCategoryAlgorithmPlugin && len(appInstance.Spec.Modules) > 0 {
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

			for index, sdk := range host.Spec.Sdks {
				if sdk.AppRef.Name == appInstance.Spec.AppRef.Name && sdk.AppRef.Version == appInstance.Spec.AppRef.Version {
					host.Spec.Sdks = append(host.Spec.Sdks[:index], host.Spec.Sdks[index+1:]...)
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

func NewAppInstance() *AppInstance {
	appInstance := new(AppInstance)
	appInstance.Init(ApiVersion, core.KindAppInstance)
	appInstance.Spec.LivenessProbe.InitialDelaySeconds = 10
	appInstance.Spec.LivenessProbe.PeriodSeconds = 60
	appInstance.Spec.LivenessProbe.TimeoutSeconds = 60
	appInstance.Spec.Modules = []AppInstanceModule{}
	return appInstance
}

func NewAppInstanceRegistry() AppInstanceRegistry {
	r := AppInstanceRegistry{
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
	r.SetPostDeleteHook(appInstancePostDelete)
	return r
}
