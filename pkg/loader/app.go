package loader

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/setting"
)

const (
	PackageConfDir = "data/binaries-conf"
	BaseAppConfDir = "manifests/init-data"
)

type Apps []App

type App struct {
	Name              string        `yaml:"name"`
	ShortName         string        `yaml:"short_name"`
	Desc              string        `yaml:"desc"`
	Increment         bool          `yaml:"increment"`
	Version           string        `yaml:"version"`
	Platform          string        `yaml:"platform"`
	Category          string        `yaml:"category"`
	SupportActions    []string      `yaml:"support_actions"`
	SupportMediaTypes []string      `yaml:"support_media_types"`
	SupportGpuModels  []string      `yaml:"support_gpu_models"`
	PkgRef            string        `yaml:"pkg_ref"`
	DashboardId       string        `yaml:"dashboardid"`
	LivenessProbe     LivenessProbe `yaml:"liveness_probe"`
	Modules           []Module      `yaml:"modules"`
	Global            Global        `yaml:"global"`
}

type LivenessProbe struct {
	InitialDelaySeconds int `yaml:"initial_delay_seconds"`
	PeriodSeconds       int `yaml:"period_seconds"`
	TimeoutSeconds      int `yaml:"timeout_seconds"`
}

type Global struct {
	Args        []Arg    `yaml:"args"`
	Configs     []string `yaml:"configs"`
	HostAliases []string `yaml:"host_aliases"`
}

type Module struct {
	Name              string                 `yaml:"name"`
	Desc              string                 `yaml:"desc"`
	SkipUpgrade       bool                   `yaml:"skip_upgrade"`
	Notes             string                 `yaml:"notes"`
	Args              []Arg                  `yaml:"args"`
	Configs           []string               `yaml:"configs"`
	Required          bool                   `yaml:"required"`
	EnableLogging     bool                   `yaml:"enable_logging"`
	EnablePurgeData   bool                   `yaml:"enable_purge_data"`
	Replication       bool                   `yaml:"replication"`
	Resources         Resources              `yaml:"resources"`
	HostLimits        HostLimits             `yaml:"host_limits"`
	IncludeRoles      []string               `yaml:"include_roles"`
	ExtraVars         map[string]interface{} `yaml:"extra_vars"`
	HostAliases       []string               `yaml:"host_aliases"`
	AdditionalConfigs v1.AdditionalConfigs   `yaml:"additional_configs"`
}

type Resources struct {
	AlgorithPlugin                bool   `yaml:"algorithm_plugin"`
	SupportAlgorithmPluginsRegexp string `yaml:"support_algorithm_plugins_regexp"`
}

type HostLimits struct {
	Max int `yaml:"max"`
	Min int `yaml:"min"`
}

type Arg struct {
	Name       string      `yaml:"name"`
	ShortName  string      `yaml:"short_name"`
	Desc       string      `yaml:"desc"`
	Type       string      `yaml:"type"`
	Format     string      `yaml:"format"`
	Enum       []string    `yaml:"enum"`
	HostLimits HostLimits  `yaml:"host_limits"`
	Default    interface{} `yaml:"default"`
	Required   bool        `yaml:"required"`
	Modifiable bool        `yaml:"modifiable"`
	Readonly   bool        `yaml:"readonly"`
}

func convert(m map[interface{}]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range m {
		switch v2 := v.(type) {
		case map[interface{}]interface{}:
			res[fmt.Sprint(k)] = convert(v2)
		default:
			res[fmt.Sprint(k)] = v
		}
	}
	return res
}

type Categories []string

func (s Categories) contains(item string) bool {
	for _, i := range s {
		if i == item {
			return true
		}
	}
	return false
}

// LoadApps 读取指定路径上的Apps.yml并生成app信息保存到数据库中, 加载应用时会将所有相同Category的应用先置为Enabled: false, 再根据Apps.yml存在的应用置为Enabled: true, 因此每次执行时会更新所有相同Category的应用
func LoadApps(categories Categories, path string) error {
	log.Debugf("reload %v app from %s", categories, path)

	appsHash := make(map[string]string)

	// 读取Apps.yml内容
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
		return err
	}
	apps := make([]App, 0)
	if err := yaml.Unmarshal(data, &apps); err != nil {
		log.Error(err)
		return err
	}
	helper := orm.GetHelper()

	// 获取现有的app列表，并禁用其中的自定义app，在后续遍历pkg的时候再将匹配到的app启用
	appObjs, err := helper.V1.App.List(context.TODO(), core.DefaultNamespace)
	if err != nil {
		log.Error(err)
		return err
	}
	for _, appObj := range appObjs {
		app := appObj.(*v1.App)
		if !categories.contains(app.Spec.Category) {
			continue
		}
		// 记录应用更新前的哈希值
		appsHash[app.Metadata.Name] = app.SpecHash()
		// 将app中的所有版本可用性都置为false
		for versionAppIndex, _ := range app.Spec.Versions {
			app.Spec.Versions[versionAppIndex].Enabled = false
		}
	}

	for _, app := range apps {
		algorithmPluginType := ""

		if !categories.contains(app.Category) {
			continue
		}

		// 待创建的配置文件
		configs := []*v1.ConfigMap{}

		// 填充模块配置
		modules := []v1.AppModule{}
		for _, module := range app.Modules {
			// 填充应用模块参数
			args := []v1.AppArgs{}
			for _, arg := range module.Args {
				args = append(args, v1.AppArgs{
					Name:      arg.Name,
					ShortName: arg.ShortName,
					Desc:      arg.Desc,
					Type:      arg.Type,
					Format:    arg.Format,
					Enum:      arg.Enum,
					Default:   arg.Default,
					Readonly:  arg.Readonly,
					HostLimits: v1.HostLimits{
						Max: arg.HostLimits.Max,
						Min: arg.HostLimits.Min,
					},
					Required:   arg.Required,
					Modifiable: arg.Modifiable,
				})
			}

			// 填充自定义配置模块参数
			additional_configmap_args := []v1.AppArgs{}
			for _, arg := range module.AdditionalConfigs.Args {
				additional_configmap_args = append(additional_configmap_args, v1.AppArgs{
					Name:      arg.Name,
					ShortName: arg.ShortName,
					Desc:      arg.Desc,
					Type:      arg.Type,
					Format:    arg.Format,
					Enum:      arg.Enum,
					Default:   arg.Default,
					Readonly:  arg.Readonly,
					HostLimits: v1.HostLimits{
						Max: arg.HostLimits.Max,
						Min: arg.HostLimits.Min,
					},
					Required:   arg.Required,
					Modifiable: arg.Modifiable,
				})

			}

			// 填充应用模块配置文件
			configMapRef := v1.ConfigMapRef{}
			if module.Configs != nil && len(module.Configs) > 0 {
				configMap := v1.NewConfigMap()
				configMap.Metadata.Namespace = core.DefaultNamespace
				configMap.Metadata.Name = "configs-app-" + app.Name + "-" + app.Version + "-" + module.Name
				for _, config := range module.Configs {
					if app.Category == core.AppCategoryCustomize {
						data, err := ioutil.ReadFile(filepath.Join(setting.PackageSetting.PkgPath, app.PkgRef, PackageConfDir, module.Name, config))
						if err != nil {
							log.Error(err)
							continue
						}
						configMap.Data[config] = string(data)
					} else if app.Category == core.AppCategoryThirdParty {
						data, err := ioutil.ReadFile(filepath.Join(setting.AnsibleSetting.BaseDir, BaseAppConfDir, module.Name, config))
						if err != nil {
							log.Error(err)
							continue
						}
						configMap.Data[config] = string(data)
					}

				}
				configs = append(configs, configMap)

				configMapRef.Namespace = configMap.Metadata.Namespace
				configMapRef.Name = configMap.Metadata.Name
			}

			extraVars := make(map[string]interface{})
			for varName, varValue := range module.ExtraVars {
				switch v := varValue.(type) {
				case map[interface{}]interface{}:
					extraVars[varName] = convert(v)
				default:
					extraVars[varName] = v
				}
				if varName == "ap_type" {
					algorithmPluginType, _ = varValue.(string)
				}
			}

			modules = append(modules, v1.AppModule{
				Name:        module.Name,
				Desc:        module.Desc,
				Notes:       module.Notes,
				SkipUpgrade: module.SkipUpgrade,
				HostLimits: v1.HostLimits{
					Max: module.HostLimits.Max,
					Min: module.HostLimits.Min,
				},
				Resources: v1.Resources{
					AlgorithmPlugin:               module.Resources.AlgorithPlugin,
					SupportAlgorithmPluginsRegexp: module.Resources.SupportAlgorithmPluginsRegexp,
				},
				IncludeRoles:    module.IncludeRoles,
				Required:        module.Required,
				EnableLogging:   module.EnableLogging,
				EnablePurgeData: module.EnablePurgeData,
				Replication:     module.Replication,
				Args:            args,
				ConfigMapRef:    configMapRef,
				ExtraVars:       extraVars,
				HostAliases:     module.HostAliases[:],
				AdditionalConfigs: v1.AdditionalConfigs{
					Enabled:      module.AdditionalConfigs.Enabled,
					ConfigMapRef: configMapRef,
					Args:         additional_configmap_args,
				},
			})
		}

		// 填充全局参数
		globalArgs := []v1.AppArgs{}
		for _, arg := range app.Global.Args {
			globalArgs = append(globalArgs, v1.AppArgs{
				Name:      arg.Name,
				ShortName: arg.ShortName,
				Desc:      arg.Desc,
				Type:      arg.Type,
				Format:    arg.Format,
				Enum:      arg.Enum,
				Default:   arg.Default,
				Readonly:  arg.Readonly,
				HostLimits: v1.HostLimits{
					Max: arg.HostLimits.Max,
					Min: arg.HostLimits.Min,
				},
				Required:   arg.Required,
				Modifiable: arg.Modifiable,
			})
		}

		// 填充全局配置文件
		globalConfigMapRef := v1.ConfigMapRef{}
		if app.Global.Configs != nil && len(app.Global.Configs) > 0 {
			globalConfigMap := v1.NewConfigMap()
			globalConfigMap.Metadata.Namespace = core.DefaultNamespace
			globalConfigMap.Metadata.Name = "configs-app-" + app.Name + "-" + app.Version + "-global"
			for _, config := range app.Global.Configs {
				if app.Category == core.AppCategoryCustomize {
					data, err := ioutil.ReadFile(filepath.Join(setting.PackageSetting.PkgPath, PackageConfDir, config))
					if err != nil {
						log.Error(err)
						continue
					}
					globalConfigMap.Data[config] = string(data)
				} else if app.Category == core.AppCategoryThirdParty {
					data, err := ioutil.ReadFile(filepath.Join(setting.PackageSetting.PkgPath, PackageConfDir, config))
					if err != nil {
						log.Error(err)
						continue
					}
					globalConfigMap.Data[config] = string(data)
				}
			}
			configs = append(configs, globalConfigMap)

			globalConfigMapRef.Namespace = globalConfigMapRef.Namespace
			globalConfigMapRef.Name = globalConfigMapRef.Name
		}

		// 填充版本应用
		versionApp := v1.AppVersion{
			Platform:          app.Platform,
			Version:           app.Version,
			ShortName:         app.ShortName,
			Desc:              app.Desc,
			SupportActions:    app.SupportActions,
			SupportGpuModels:  app.SupportGpuModels,
			SupportMediaTypes: app.SupportMediaTypes,
			Increment:         app.Increment,
			Enabled:           true,
			DashboardId:       app.DashboardId,
			PkgRef:            app.PkgRef,
			LivenessProbe: v1.LivenessProbe{
				InitialDelaySeconds: app.LivenessProbe.InitialDelaySeconds,
				PeriodSeconds:       app.LivenessProbe.PeriodSeconds,
				TimeoutSeconds:      app.LivenessProbe.TimeoutSeconds,
			},
			Modules: modules,
			Global: v1.AppGlobal{
				Args:         globalArgs,
				ConfigMapRef: globalConfigMapRef,
				HostAliases:  app.Global.HostAliases[:],
			},
		}

		// 用于标记应用是否已存在
		appExist := false

		for _, appObj := range appObjs {
			oriApp := appObj.(*v1.App)
			if oriApp.Metadata.Name == app.Name {
				oriApp.Spec.Category = app.Category
				oriApp.Spec.Platform = app.Platform

				// 用于标记应用版本是否已存在
				versionAppExist := false

				for versionAppIndex, oriVersionApp := range oriApp.Spec.Versions {
					// 如果应用版本已存在, 则覆盖旧版本
					if oriVersionApp.Version == app.Version {
						oriApp.Spec.Versions[versionAppIndex] = versionApp

						versionAppExist = true
						break
					}
				}

				// 如果应用对应的版本不存在则追加一个新的版本，并更新应用
				if !versionAppExist {
					log.Infof("append app %s with new version %s", app.Name, app.Version)

					oriApp.Spec.Versions = append(oriApp.Spec.Versions, versionApp)
				}

				oriApp.Metadata.Annotations["ShortName"] = app.ShortName

				appExist = true
				break
			}
		}

		// 如果应用不存在则创建一个新的应用
		if !appExist {
			log.Infof("append new app %s with version %s", app.Name, app.Version)

			newApp := v1.NewApp()
			newApp.Metadata.Namespace = core.DefaultNamespace
			newApp.Metadata.Name = app.Name
			newApp.Metadata.Annotations["ShortName"] = app.ShortName
			if app.Category == core.AppCategoryAlgorithmPlugin {
				newApp.Metadata.Annotations[core.AnnotationAlgorithmPluginPrefix+"type"] = algorithmPluginType
			}
			newApp.Spec.Category = app.Category
			newApp.Spec.Category = app.Platform
			newApp.Spec.Versions = append(newApp.Spec.Versions, versionApp)

			if _, err := helper.V1.App.Create(context.TODO(), newApp); err != nil {
				log.Error(err)
				return err
			}
			appObjs = append(appObjs, newApp)
		}

		// 创建或更新应用配置文件
		for _, config := range configs {
			if cm, err := helper.V1.ConfigMap.Get(context.TODO(), config.Metadata.Namespace, config.Metadata.Name); err != nil {
				log.Error(err)
				return err
			} else if cm != nil {
				if cm.SpecHash() != config.SpecHash() {
					if _, err := helper.V1.ConfigMap.Update(context.TODO(), config); err != nil {
						log.Error(err)
						return err
					}
				}
			} else {
				if _, err := helper.V1.ConfigMap.Create(context.TODO(), config); err != nil {
					log.Error(err)
					return err
				}
			}
		}
	}

	// 对于已存在的应用进行更新
	for _, appObj := range appObjs {
		app := appObj.(*v1.App)
		if !categories.contains(app.Spec.Category) || appsHash[app.Metadata.Name] == app.SpecHash() {
			continue
		}
		if _, err := helper.V1.App.Update(context.TODO(), app, core.WhenSpecChanged()); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}
