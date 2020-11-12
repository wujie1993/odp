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
