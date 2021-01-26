package wavectl

import (
	"context"

	log "github.com/sirupsen/logrus"

	clientset "github.com/wujie1993/waves/pkg/client"
	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v2"
)

const (
	HostPluginActionInstall   = "install"
	HostPluginActionUninstall = "uninstall"
)

type HostPluginClient struct {
	clientset.ClientSet
}

// Install 安装主机插件
func (c HostPluginClient) Install(hostName string, pluginName string, pluginVersion string, force bool) error {
	host, err := c.ClientSet.V2().Hosts().Get(context.TODO(), hostName)
	if err != nil {
		log.Error(err)
		return err
	} else if host == nil {
		err := e.Errorf("host %s not found", hostName)
		log.Error(err)
		return err
	}

	app, err := c.ClientSet.V1().Apps(core.DefaultNamespace).Get(context.TODO(), pluginName)
	if err != nil {
		log.Error(err)
		return err
	} else if app == nil {
		err := e.Errorf("host plugin %s not found", pluginName)
		log.Error(err)
		return err
	}

	versionedPlugin, ok := app.GetVersion(pluginVersion)
	if !ok {
		err := e.Errorf("host plugin %s does not contain with version %s", pluginName, pluginVersion)
		log.Error(err)
		return err
	}
	var appInstance *v2.AppInstance
	var pluginExist bool
	for _, plugin := range host.Info.Plugins {
		if plugin.AppRef.Name == pluginName {
			appInstance, err = c.ClientSet.V2().AppInstances(plugin.AppInstanceRef.Namespace).Get(context.TODO(), plugin.AppInstanceRef.Name)
			if err != nil {
				log.Error(err)
				return err
			} else if appInstance == nil {
				continue
			} else if appInstance.Status.Phase != core.PhaseUninstalled && !force {
				err := e.Errorf("host plugin %s is %s", pluginName, appInstance.Status.Phase)
				log.Error(err)
				return err
			}
			pluginExist = true
			break
		}
	}

	if appInstance == nil {
		appInstance = v2.NewAppInstance()
		appInstance.Metadata.Namespace = core.DefaultNamespace
		appInstance.Metadata.Name = "host-" + hostName + "-plugin-" + pluginName
	}
	appInstance.Spec.Category = core.AppCategoryHostPlugin
	appInstance.Spec.Action = core.AppActionInstall
	appInstance.Spec.AppRef = v2.AppRef{
		Name:    pluginName,
		Version: pluginVersion,
	}
	appInstance.Spec.Modules = []v2.AppInstanceModule{}
	for _, appModule := range versionedPlugin.Modules {
		moduleReplica := v2.AppInstanceModuleReplica{
			HostRefs: []string{hostName},
			Args:     []v2.AppInstanceArgs{},
		}
		for _, arg := range appModule.Args {
			moduleReplica.Args = append(moduleReplica.Args, v2.AppInstanceArgs{
				Name:  arg.Name,
				Value: arg.Default,
			})
		}
		appInstance.Spec.Modules = append(appInstance.Spec.Modules, v2.AppInstanceModule{
			Name:       appModule.Name,
			AppVersion: pluginVersion,
			Replicas:   []v2.AppInstanceModuleReplica{moduleReplica},
		})
	}
	for _, arg := range versionedPlugin.Global.Args {
		appInstance.Spec.Global.Args = append(appInstance.Spec.Global.Args, v2.AppInstanceArgs{
			Name:  arg.Name,
			Value: arg.Default,
		})
	}

	if !pluginExist {
		if _, err := c.ClientSet.V2().AppInstances(appInstance.Metadata.Namespace).Create(context.TODO(), appInstance); err != nil {
			log.Error(err)
			return err
		}
	} else {
		if _, err := c.ClientSet.V2().AppInstances(appInstance.Metadata.Namespace).Update(context.TODO(), appInstance); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

// Install 卸载主机插件
func (c HostPluginClient) Uninstall(hostName string, pluginName string, force bool) error {
	host, err := c.ClientSet.V2().Hosts().Get(context.TODO(), hostName)
	if err != nil {
		log.Error(err)
		return err
	} else if host == nil {
		err := e.Errorf("host %s not found", hostName)
		log.Error(err)
		return err
	}

	app, err := c.ClientSet.V1().Apps(core.DefaultNamespace).Get(context.TODO(), pluginName)
	if err != nil {
		log.Error(err)
		return err
	} else if app == nil {
		err := e.Errorf("host plugin %s not found", pluginName)
		log.Error(err)
		return err
	}

	var appInstance *v2.AppInstance
	for _, plugin := range host.Info.Plugins {
		if plugin.AppRef.Name == pluginName {
			appInstance, err = c.ClientSet.V2().AppInstances(core.DefaultNamespace).Get(context.TODO(), plugin.AppInstanceRef.Name)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	}
	if appInstance == nil {
		err := e.Errorf("host %s does not installed with plugin %s", hostName, pluginName)
		log.Error(err)
		return err
	}
	if !force {
		if appInstance.Status.Phase != core.PhaseInstalled && appInstance.Status.Phase != core.PhaseFailed {
			err := e.Errorf("host plugin %s is %s", pluginName, appInstance.Status.Phase)
			log.Error(err)
			return err
		}
	}

	appInstance.Spec.Action = core.AppActionUninstall
	if _, err := c.ClientSet.V2().AppInstances(appInstance.Metadata.Namespace).Update(context.TODO(), appInstance); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
