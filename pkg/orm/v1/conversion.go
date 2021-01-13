package v1

import (
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/runtime"
)

var conversion core.Conversion

func init() {
	conversion = core.NewConversion()

	// 注册v1版本结构与运行时结构互相转换方法
	registerConversionFunc(core.VK{
		ApiVersion: ApiVersion,
		Kind:       core.KindAppInstance,
	}, core.VK{
		Kind: core.KindAppInstance,
	}, convertCoreV1AppInstanceToCoreRuntimeAppInstance)

	registerConversionFunc(core.VK{
		Kind: core.KindAppInstance,
	}, core.VK{
		ApiVersion: ApiVersion,
		Kind:       core.KindAppInstance,
	}, convertCoreRuntimeAppInstanceToCoreV1AppInstance)

	registerConversionFunc(core.VK{
		ApiVersion: ApiVersion,
		Kind:       core.KindJob,
	}, core.VK{
		Kind: core.KindJob,
	}, convertCoreV1JobToCoreRuntimeJob)

	registerConversionFunc(core.VK{
		Kind: core.KindJob,
	}, core.VK{
		ApiVersion: ApiVersion,
		Kind:       core.KindJob,
	}, convertCoreRuntimeJobToCoreV1Job)

	registerConversionFunc(core.VK{
		ApiVersion: ApiVersion,
		Kind:       core.KindHost,
	}, core.VK{
		Kind: core.KindHost,
	}, convertCoreV1HostToCoreRuntimeHost)

	registerConversionFunc(core.VK{
		Kind: core.KindHost,
	}, core.VK{
		ApiVersion: ApiVersion,
		Kind:       core.KindHost,
	}, convertCoreRuntimeHostToCoreV1Host)
}

func registerConversionFunc(srcVK core.VK, dstVK core.VK, convertFunc core.ConvertFunc) {
	conversion.SetConversionFunc(core.GVK{
		Group:      core.Group,
		ApiVersion: srcVK.ApiVersion,
		Kind:       srcVK.Kind,
	}, core.GVK{
		Group:      core.Group,
		ApiVersion: dstVK.ApiVersion,
		Kind:       dstVK.Kind,
	}, convertFunc)
}

// Convert 将v1版本结构与运行时结构互相转换
func Convert(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
	srcGVK := srcObj.GetGVK()

	if srcGVK == dstGVK {
		// 源与目标结构一致，直接返回源目标对象
		return srcObj, nil
	}

	if (srcGVK.ApiVersion == "" && dstGVK.ApiVersion == "") || (srcGVK.ApiVersion != "" && dstGVK.ApiVersion != "") {
		return nil, e.Errorf("Convert %s to %+v failed. Unsupported conversion", srcObj.GetKey(), dstGVK)
	}

	log.Tracef("convert %+v %+v from %+v to %+v", reflect.TypeOf(srcObj), srcObj, srcGVK, dstGVK)
	// 直接转换
	convertFunc, ok := conversion.GetConversionFunc(srcGVK, dstGVK)
	if !ok {
		return nil, e.Errorf("Convert %s to %+v failed. Convert function not found", srcObj.GetKey(), dstGVK)
	}
	return convertFunc(srcObj, dstGVK)
}

func convertCoreV1AppInstanceToCoreRuntimeAppInstance(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
	srcAppInstance, ok := srcObj.(*AppInstance)
	if !ok {
		return nil, e.Errorf("mismatch with type of source object")
	}
	dstObj, err := runtime.New(dstGVK.Kind)
	if err != nil {
		return nil, err
	}
	if err := core.DeepCopy(srcObj, dstObj); err != nil {
		return nil, err
	}
	dstAppInstance, ok := dstObj.(*runtime.AppInstance)
	if !ok {
		return nil, e.Errorf("mismatch with type of destnation object")
	}
	for _, srcModule := range srcAppInstance.Spec.Modules {
		for dstModuleIndex, dstModule := range dstAppInstance.Spec.Modules {
			if srcModule.Name == dstModule.Name {
				moduleReplica := runtime.AppInstanceModuleReplica{}
				core.DeepCopy(srcModule, &moduleReplica)
				dstAppInstance.Spec.Modules[dstModuleIndex].Replicas = []runtime.AppInstanceModuleReplica{moduleReplica}
			}
		}
	}
	dstAppInstance.SetGVK(dstGVK)
	return dstAppInstance, nil
}

func convertCoreRuntimeAppInstanceToCoreV1AppInstance(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
	srcAppInstance, ok := srcObj.(*runtime.AppInstance)
	if !ok {
		return nil, e.Errorf("mismatch with type of source object")
	}
	dstObj, err := New(dstGVK.Kind)
	if err != nil {
		return nil, err
	}
	if err := core.DeepCopy(srcObj, dstObj); err != nil {
		return nil, err
	}
	dstAppInstance, ok := dstObj.(*AppInstance)
	if !ok {
		return nil, e.Errorf("mismatch with type of destnation object")
	}
	for _, srcModule := range srcAppInstance.Spec.Modules {
		for dstModuleIndex, dstModule := range dstAppInstance.Spec.Modules {
			if srcModule.Name == dstModule.Name && len(srcModule.Replicas) > 0 {
				if err := core.DeepCopy(srcModule.Replicas[0], &dstAppInstance.Spec.Modules[dstModuleIndex]); err != nil {
					return nil, err
				}
			}
		}
	}
	dstAppInstance.SetGVK(dstGVK)
	return dstAppInstance, nil
}

func convertCoreV1JobToCoreRuntimeJob(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
	srcJob, ok := srcObj.(*Job)
	if !ok {
		return nil, e.Errorf("mismatch with type of source object")
	}
	dstObj, err := runtime.New(dstGVK.Kind)
	if err != nil {
		return nil, err
	}
	if err := core.DeepCopy(srcObj, dstObj); err != nil {
		return nil, err
	}
	dstJob, ok := dstObj.(*runtime.Job)
	if !ok {
		return nil, e.Errorf("mismatch with type of destnation object")
	}
	play := runtime.JobAnsiblePlay{
		Name:    "default",
		Configs: []runtime.AnsibleConfig{},
		Envs:    srcJob.Spec.Exec.Ansible.Envs,
		Tags:    srcJob.Spec.Exec.Ansible.Tags,
	}
	for _, config := range srcJob.Spec.Exec.Ansible.Configs {
		play.Configs = append(play.Configs, runtime.AnsibleConfig{
			PathPrefix: config.Path,
			ValueFrom: runtime.ValueFrom{
				ConfigMapRef: runtime.ConfigMapRef{
					Namespace: config.ConfigMapRef.Namespace,
					Name:      config.ConfigMapRef.Name,
				},
			},
		})
	}
	play.GroupVars = runtime.AnsibleGroupVars{
		ValueFrom: runtime.ValueFrom{
			ConfigMapRef: runtime.ConfigMapRef{
				Namespace: srcJob.Spec.Exec.Ansible.GroupVars.ValueFrom.ConfigMapRef.Namespace,
				Name:      srcJob.Spec.Exec.Ansible.GroupVars.ValueFrom.ConfigMapRef.Name,
			},
		},
	}
	if len(srcJob.Spec.Exec.Ansible.Inventories) > 0 {
		inventoryValue := ""
		for _, inventory := range srcJob.Spec.Exec.Ansible.Inventories {
			inventoryValue += inventory.Value + "\n"
		}
		play.Inventory = runtime.AnsibleInventory{
			Value: inventoryValue,
			ValueFrom: runtime.ValueFrom{
				ConfigMapRef: runtime.ConfigMapRef{
					Namespace: srcJob.Spec.Exec.Ansible.Inventories[0].ValueFrom.ConfigMapRef.Namespace,
					Name:      srcJob.Spec.Exec.Ansible.Inventories[0].ValueFrom.ConfigMapRef.Name,
				},
			},
		}
	}
	play.Playbook = runtime.AnsiblePlaybook{
		Value: srcJob.Spec.Exec.Ansible.Playbook,
	}
	dstJob.Spec.Exec.Ansible.Plays = []runtime.JobAnsiblePlay{play}
	dstObj.SetGVK(dstGVK)
	return dstObj, nil
}

func convertCoreRuntimeJobToCoreV1Job(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
	srcJob, ok := srcObj.(*runtime.Job)
	if !ok {
		return nil, e.Errorf("mismatch with type of source object")
	}
	dstObj, err := New(dstGVK.Kind)
	if err != nil {
		return nil, err
	}
	if err := core.DeepCopy(srcObj, dstObj); err != nil {
		return nil, err
	}
	dstJob, ok := dstObj.(*Job)
	if !ok {
		return nil, e.Errorf("mismatch with type of destnation object")
	}
	if len(srcJob.Spec.Exec.Ansible.Plays) > 0 {
		play := srcJob.Spec.Exec.Ansible.Plays[0]
		dstJob.Spec.Exec.Ansible.Envs = play.Envs
		dstJob.Spec.Exec.Ansible.Tags = play.Tags
		configs := []JobConfig{}
		for _, config := range play.Configs {
			configs = append(configs, JobConfig{
				Path: config.PathPrefix,
				ConfigMapRef: ConfigMapRef{
					Namespace: config.ValueFrom.ConfigMapRef.Namespace,
					Name:      config.ValueFrom.ConfigMapRef.Name,
				},
			})
		}
		dstJob.Spec.Exec.Ansible.Configs = configs
		dstJob.Spec.Exec.Ansible.GroupVars = GroupVars{
			ValueFrom: ValueFrom{
				ConfigMapRef: ConfigMapRef{
					Namespace: play.GroupVars.ValueFrom.ConfigMapRef.Namespace,
					Name:      play.GroupVars.ValueFrom.ConfigMapRef.Name,
				},
			},
		}
		dstJob.Spec.Exec.Ansible.Inventories = []AnsibleInventory{
			{
				Value: play.Inventory.Value,
				ValueFrom: ValueFrom{
					ConfigMapRef: ConfigMapRef{
						Namespace: play.Inventory.ValueFrom.ConfigMapRef.Namespace,
						Name:      play.Inventory.ValueFrom.ConfigMapRef.Name,
					},
				},
			},
		}
		dstJob.Spec.Exec.Ansible.Playbook = play.Playbook.Value
	}
	dstObj.SetGVK(dstGVK)
	return dstObj, nil
}

func convertCoreV1HostToCoreRuntimeHost(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
	srcJob, ok := srcObj.(*Host)
	if !ok {
		return nil, e.Errorf("mismatch with type of source object")
	}
	dstObj, err := runtime.New(dstGVK.Kind)
	if err != nil {
		return nil, err
	}
	dstObj.SetMetadata(srcObj.GetMetadata())
	dstObj.SetStatus(srcObj.GetStatus())

	dstJob, ok := dstObj.(*runtime.Host)
	if !ok {
		return nil, e.Errorf("mismatch with type of destnation object")
	}
	dstJob.Spec.SSH = runtime.HostSSH{
		Host:     srcJob.Spec.SSH.Host,
		User:     srcJob.Spec.SSH.User,
		Password: srcJob.Spec.SSH.Password,
		Port:     srcJob.Spec.SSH.Port,
	}
	dstJob.Info = runtime.HostInfo{
		OS: runtime.OS{
			Kernel:  srcJob.Spec.Info.OS.Kernel,
			Release: srcJob.Spec.Info.OS.Release,
		},
		CPU: runtime.CPU{
			Cores: srcJob.Spec.Info.CPU.Cores,
			Model: srcJob.Spec.Info.CPU.Model,
		},
		Disk: runtime.Disk{
			Size: srcJob.Spec.Info.Disk.Size,
		},
		Memory: runtime.Memory{
			Size:  srcJob.Spec.Info.Memory.Size,
			Model: srcJob.Spec.Info.Memory.Model,
		},
	}
	dstJob.Info.GPUs = []runtime.GPUInfo{}
	for _, gpu := range srcJob.Spec.Info.GPUs {
		dstJob.Info.GPUs = append(dstJob.Info.GPUs, runtime.GPUInfo{
			ID:     gpu.ID,
			Memory: gpu.Memory,
			Model:  gpu.Model,
			Type:   gpu.Type,
			UUID:   gpu.UUID,
		})
	}
	dstJob.Info.Plugins = []runtime.HostPlugin{}
	for _, plugin := range srcJob.Spec.Plugins {
		dstJob.Info.Plugins = append(dstJob.Info.Plugins, runtime.HostPlugin{
			AppInstanceRef: runtime.AppInstanceRef{
				Namespace: plugin.AppInstanceRef.Namespace,
				Name:      plugin.AppInstanceRef.Name,
			},
			AppRef: runtime.AppRef{
				Version: plugin.AppRef.Version,
				Name:    plugin.AppRef.Name,
			},
		})
	}
	dstObj.SetGVK(dstGVK)
	return dstObj, nil
}

func convertCoreRuntimeHostToCoreV1Host(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
	srcJob, ok := srcObj.(*runtime.Host)
	if !ok {
		return nil, e.Errorf("mismatch with type of source object")
	}
	dstObj, err := New(dstGVK.Kind)
	if err != nil {
		return nil, err
	}
	dstObj.SetMetadata(srcObj.GetMetadata())
	dstObj.SetStatus(srcObj.GetStatus())

	dstJob, ok := dstObj.(*Host)
	if !ok {
		return nil, e.Errorf("mismatch with type of destnation object")
	}
	dstJob.Spec.SSH = HostSSH{
		Host:     srcJob.Spec.SSH.Host,
		User:     srcJob.Spec.SSH.User,
		Password: srcJob.Spec.SSH.Password,
		Port:     srcJob.Spec.SSH.Port,
	}
	dstJob.Spec.Info = HostInfo{
		OS: OS{
			Kernel:  srcJob.Info.OS.Kernel,
			Release: srcJob.Info.OS.Release,
		},
		CPU: CPU{
			Cores: srcJob.Info.CPU.Cores,
			Model: srcJob.Info.CPU.Model,
		},
		Disk: Disk{
			Size: srcJob.Info.Disk.Size,
		},
		Memory: Memory{
			Size:  srcJob.Info.Memory.Size,
			Model: srcJob.Info.Memory.Model,
		},
	}
	dstJob.Spec.Info.GPUs = []GPUInfo{}
	for _, gpu := range srcJob.Info.GPUs {
		dstJob.Spec.Info.GPUs = append(dstJob.Spec.Info.GPUs, GPUInfo{
			ID:     gpu.ID,
			Memory: gpu.Memory,
			Model:  gpu.Model,
			Type:   gpu.Type,
			UUID:   gpu.UUID,
		})
	}
	dstJob.Spec.Plugins = []HostPlugin{}
	for _, plugin := range srcJob.Info.Plugins {
		dstJob.Spec.Plugins = append(dstJob.Spec.Plugins, HostPlugin{
			AppInstanceRef: AppInstanceRef{
				Namespace: plugin.AppInstanceRef.Namespace,
				Name:      plugin.AppInstanceRef.Name,
			},
			AppRef: AppRef{
				Version: plugin.AppRef.Version,
				Name:    plugin.AppRef.Name,
			},
		})
	}
	dstObj.SetGVK(dstGVK)
	return dstObj, nil
}
