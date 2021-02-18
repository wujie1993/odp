package operators

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/ansible"
	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/setting"
	"github.com/wujie1993/waves/pkg/util"
)

// HostOperator 主机管理器
type HostOperator struct {
	BaseOperator
}

// handleHost 处理主机的变更操作
func (c *HostOperator) handleHost(ctx context.Context, obj core.ApiObject) error {
	host := obj.(*v1.Host)
	log.Infof("%s '%s' is %s", host.Kind, host.GetKey(), host.Status.Phase)

	switch host.Status.Phase {
	case core.PhaseWaiting:
		// 处于等待中状态的主机会将状态更新为连接中
		if _, err := c.helper.V1.Host.UpdateStatusPhase(host.Metadata.Namespace, host.Metadata.Name, core.PhaseConnecting); err != nil {
			log.Error(err)
			return err
		}
	case core.PhaseConnecting:
		// 处于连接中状态的主机会首先尝试ssh连接并获取主机状态，并将状态更新为初始化中

		info, err := util.GetMachineInfo(
			host.Spec.SSH.Host,
			host.Spec.SSH.User,
			host.Spec.SSH.Password,
			host.Spec.SSH.Port,
		)
		if err != nil {
			// 信息收集失败，将状态置为未就绪
			log.Errorf("failed to connect host %s: %s", host.GetKey(), err)

			c.failback(host, core.EventActionConnect, err.Error(), nil)
			return err
		}

		// 将主机状态置为初始化中
		// 获取gpu信息
		host.Spec.Info.GPUs = []v1.GPUInfo{}
		gpuinfo, err := util.GetGPUInfo(
			host.Spec.SSH.Host,
			host.Spec.SSH.User,
			host.Spec.SSH.Password,
			host.Spec.SSH.Port,
		)
		if err != nil {
			log.Error(err)
		}
		for slotIndex, record := range gpuinfo {
			gpuInfo := v1.GPUInfo{
				ID:     record.ID,
				UUID:   record.UUID,
				Model:  record.Model,
				Memory: record.Memory,
				Type:   setting.GetGPUType(record.Model),
			}
			host.Spec.Info.GPUs = append(host.Spec.Info.GPUs, gpuInfo)
			gpuObj, err := c.helper.V1.GPU.Get(context.TODO(), "", c.helper.V1.GPU.GetGPUName(host.Metadata.Name, slotIndex))
			if err != nil {
				log.Error(err)
				c.failback(host, core.EventActionConnect, err.Error(), nil)
				return err
			}
			if gpuObj == nil {
				// 当GPU资源不存在时，创建GPU资源
				gpu := v1.NewGPU()
				gpu.Metadata.Name = fmt.Sprintf("%s-slot-%d", host.Metadata.Name, slotIndex)
				gpu.Spec.HostRef = host.Metadata.Name
				gpu.Spec.Info = gpuInfo
				if _, err := c.helper.V1.GPU.Create(context.TODO(), gpu); err != nil {
					log.Error(err)
					c.failback(host, core.EventActionConnect, err.Error(), nil)
					return err
				}
			} else {
				// 当GPU资源存在时，更新GPU资源
				gpu := gpuObj.(*v1.GPU)
				gpu.Spec.HostRef = host.Metadata.Name
				gpu.Spec.Info = gpuInfo
				if _, err := c.helper.V1.GPU.Update(context.TODO(), gpu); err != nil {
					log.Error(err)
					c.failback(host, core.EventActionConnect, err.Error(), nil)
					return err
				}
			}
		}
		host.Spec.Info.OS.Release = info.OS
		host.Spec.Info.CPU.Cores = info.CPUCores
		host.Spec.Info.Memory.Size = info.MemorySize
		host.Spec.Info.Disk.Size = info.DiskSize
		delete(host.Metadata.Annotations, core.AnnotationJobPrefix+ansible.ANSIBLE_ROLE_HOST_INIT)
		host.Status.SetCondition(core.ConditionTypeConnected, core.ConditionStatusTrue)
		host.Status.Phase = core.PhaseInitialing
		if _, err := c.helper.V1.Host.Update(context.TODO(), host, core.WithAllFields()); err != nil {
			log.Error(err)
			return err
		}

		// 记录事件完成
		if err := c.recordEvent(Event{
			BaseApiObj: host.BaseApiObj,
			Action:     core.EventActionConnect,
			Msg:        "",
			JobRef:     "",
			Phase:      core.PhaseCompleted,
		}); err != nil {
			log.Error(err)
			return err
		}
	case core.PhaseInitialing:
		// 处于初始化状态时，创建用于初始化主机环境的任务，并在任务执行完成时将主机状态置为已就绪

		// 如果主机已经有对应的初始化任务，则侦听任务状态，在任务完成时将主机状态更新为已就绪
		if jobName, ok := host.Metadata.Annotations[core.AnnotationJobPrefix+ansible.ANSIBLE_ROLE_HOST_INIT]; ok {
			jobCtx, jobWatchCancel := context.WithCancel(ctx)
			defer jobWatchCancel()

			jobActionChan := c.helper.V1.Job.GetWatch(jobCtx, "", jobName)
			for jobAction := range jobActionChan {
				job := jobAction.Obj.(*v1.Job)
				switch jobAction.Type {
				case db.KVActionTypeDelete:
					// 如果侦听的任务被删除，移除关联的初始化任务，重新进行初始化
					if _, err := c.helper.V1.Host.UpdateStatusPhase(host.Metadata.Namespace, host.Metadata.Name, core.PhaseWaiting); err != nil {
						log.Error(err)
						return err
					}
					return nil
				case db.KVActionTypeSet:
					switch job.Status.Phase {
					case core.PhaseCompleted:
						// 如果初始化任务执行成功, 将主机状态更新为已就绪并结束任务侦听
						host.Status.SetCondition(core.ConditionTypeInitialized, core.ConditionStatusTrue)
						host.SetStatusPhase(core.PhaseReady)
						if _, err := c.helper.V1.Host.UpdateStatus(host.Metadata.Namespace, host.Metadata.Name, host.Status); err != nil {
							log.Error(err)
						}

						// 记录事件完成
						if err := c.recordEvent(Event{
							BaseApiObj: host.BaseApiObj,
							Action:     core.EventActionInitial,
							Msg:        "",
							JobRef:     job.Metadata.Name,
							Phase:      core.PhaseCompleted,
						}); err != nil {
							log.Error(err)
							return err
						}
						return nil
					case core.PhaseFailed:
						c.failback(host, core.EventActionInitial, "", job)
						return e.Errorf("host init failed")
					case core.PhaseRunning:
					default:
						log.Warnf("unknown status '%s' of job '%s'", job.Status.Phase, job.GetKey())
					}
				}
			}
			return nil
		}

		// 对于没有初始化任务的主机，创建对应的任务进行初始化并做关联

		// 模板配置解析
		inventoryTpl, err := template.New("inventory").Parse(ansible.ANSIBLE_INVENTORY_HOST_INIT_TPL)
		if err != nil {
			log.Error(err)
			return err
		}
		var inventoryBuf bytes.Buffer
		if err := inventoryTpl.Execute(&inventoryBuf, []*v1.Host{host}); err != nil {
			log.Error(err)
			return err
		}
		commonInventoryStr, err := ansible.RenderCommonInventory()
		if err != nil {
			log.Error(err)
			return err
		}

		// 创建任务对应的配置
		configMap := v1.NewConfigMap()
		configMap.Metadata.Namespace = core.DefaultNamespace
		configMap.Metadata.Name = fmt.Sprintf("%s-%s-%d", ansible.ANSIBLE_ROLE_HOST_INIT, host.Metadata.Name, time.Now().Unix())
		configMap.Data["common"] = commonInventoryStr
		configMap.Data["inventory"] = inventoryBuf.String()
		if _, err := c.helper.V1.ConfigMap.Create(context.TODO(), configMap); err != nil {
			log.Error(err)
			return err
		}

		// 生成playbook
		playbook := ansible.Playbook{
			Hosts: []string{ansible.ANSIBLE_ROLE_HOST_INIT},
			Roles: []string{ansible.ANSIBLE_ROLE_HOST_INIT},
		}
		playbookStr, err := ansible.RenderPlaybook([]ansible.Playbook{playbook})
		if err != nil {
			log.Error(err)
			return err
		}

		// 创建任务
		job := v1.NewJob()
		job.Metadata.Name = fmt.Sprintf("%s-%s-%d", ansible.ANSIBLE_ROLE_HOST_INIT, host.Metadata.Name, time.Now().Unix())
		job.Spec.Exec.Type = core.JobExecTypeAnsible
		job.Spec.Exec.Ansible.Bin = setting.AnsibleSetting.Bin
		job.Spec.Exec.Ansible.Inventories = []v1.AnsibleInventory{
			v1.AnsibleInventory{
				ValueFrom: v1.ValueFrom{
					ConfigMapRef: v1.ConfigMapRef{
						Namespace: configMap.Metadata.Namespace,
						Name:      configMap.Metadata.Name,
					},
				},
			},
		}
		job.Spec.Exec.Ansible.Envs = []string{
			"act=install",
		}
		job.Spec.Exec.Ansible.Playbook = playbookStr
		job.Spec.TimeoutSeconds = 300
		job.Spec.FailureThreshold = 3
		if _, err := c.helper.V1.Job.Create(context.TODO(), job); err != nil {
			log.Error(err)
			return err
		}

		// 将主机与任务关联
		host.Metadata.Annotations[core.AnnotationJobPrefix+ansible.ANSIBLE_ROLE_HOST_INIT] = job.Metadata.Name
		if _, err := c.helper.V1.Host.Update(context.TODO(), host); err != nil {
			log.Error(err)
			return err
		}

		// 记录事件开始
		if err := c.recordEvent(Event{
			BaseApiObj: host.BaseApiObj,
			Action:     core.EventActionInitial,
			Msg:        "",
			JobRef:     job.Metadata.Name,
			Phase:      core.PhaseWaiting,
		}); err != nil {
			log.Error(err)
		}
	case core.PhaseDeleting:
		c.delete(ctx, obj)
	}
	return nil
}

// finalizeHost 级联清除主机的关联资源
func (o HostOperator) finalizeHost(ctx context.Context, obj core.ApiObject) error {
	host := obj.(*v1.Host)

	// 每次只处理一项Finalizer
	switch host.Metadata.Finalizers[0] {
	case core.FinalizerCleanRefGPU:
		// 同步删除关联的GPU
		gpuList, err := o.helper.V1.GPU.List(context.TODO(), "")
		if err != nil {
			log.Error(err)
			return err
		}
		for _, gpuObj := range gpuList {
			gpu := gpuObj.(*v1.GPU)
			if gpu.Spec.HostRef == host.Metadata.Name {
				if _, err := o.helper.V1.GPU.Delete(context.TODO(), "", gpu.Metadata.Name, core.WithSync()); err != nil {
					log.Error(err)
					return err
				}
			}
		}
	case core.FinalizerCleanRefEvent:
		// 同步删除关联的事件
		eventList, err := o.helper.V1.Event.List(context.TODO(), "")
		if err != nil {
			log.Error(err)
			return err
		}
		for _, eventObj := range eventList {
			event := eventObj.(*v1.Event)
			if event.Spec.ResourceRef.Kind == core.KindHost && event.Spec.ResourceRef.Name == host.Metadata.Name {
				if _, err := o.helper.V1.Event.Delete(context.TODO(), "", event.Metadata.Name); err != nil {
					log.Error(err)
					return err
				}
			}
		}
	}
	return nil
}

// reconcileHost 主机定时收敛
func (o *HostOperator) reconcileHost(ctx context.Context, obj core.ApiObject) {
	host := obj.(*v1.Host)

	// 忽略非就绪状态的主机
	if host.Status.Phase != core.PhaseReady {
		return
	}

	host.Spec.Info.GPUs = []v1.GPUInfo{}
	// 获取当前的GPU信息
	gpuList, err := util.GetGPUInfo(
		host.Spec.SSH.Host,
		host.Spec.SSH.User,
		host.Spec.SSH.Password,
		host.Spec.SSH.Port,
	)
	if err != nil {
		log.Error(err)
		return
	}

	// 重置GPU记录
	gpuObjs, err := o.helper.V1.GPU.List(ctx, "")
	if err != nil {
		log.Error(err)
		return
	}
	for _, gpuObj := range gpuObjs {
		gpu := gpuObj.(*v1.GPU)
		if gpu.Spec.HostRef == host.Metadata.Name {
			gpu.Spec.Info = v1.GPUInfo{}
			if _, err := o.helper.V1.GPU.Update(context.TODO(), gpu, core.WithAllFields()); err != nil {
				log.Error(err)
				return
			}
		}
	}

	// 添加/更新GPU记录
	for slotIndex, record := range gpuList {
		gpuInfo := v1.GPUInfo{
			ID:     record.ID,
			UUID:   record.UUID,
			Model:  record.Model,
			Memory: record.Memory,
			Type:   setting.GetGPUType(record.Model),
		}
		host.Spec.Info.GPUs = append(host.Spec.Info.GPUs, gpuInfo)
		gpuObj, err := o.helper.V1.GPU.Get(context.TODO(), "", o.helper.V1.GPU.GetGPUName(host.Metadata.Name, slotIndex))
		if err != nil {
			log.Error(err)
			return
		}
		// 当GPU资源不存在，创建GPU资源
		if gpuObj == nil {
			gpu := v1.NewGPU()
			gpu.Metadata.Name = o.helper.V1.GPU.GetGPUName(host.Metadata.Name, slotIndex)
			gpu.Spec.HostRef = host.Metadata.Name
			gpu.Spec.Info = gpuInfo
			if _, err := o.helper.V1.GPU.Create(context.TODO(), gpu); err != nil {
				log.Error(err)
				return
			}
		} else {
			// 当GPU资源存在时，更新GPU资源
			gpu := gpuObj.(*v1.GPU)
			gpu.Spec.HostRef = host.Metadata.Name
			gpu.Spec.Info = gpuInfo
			if _, err := o.helper.V1.GPU.Update(context.TODO(), gpu, core.WithAllFields()); err != nil {
				log.Error(err)
				return
			}
		}
	}
	// 更新主机信息
	if _, err := o.helper.V1.Host.Update(context.TODO(), host, core.WhenSpecChanged(), core.WithAllFields()); err != nil {
		log.Error(err)
		return
	}
}

// failback 操作失败回退
func (o HostOperator) failback(obj core.ApiObject, action string, reason string, job *v1.Job) {
	host := obj.(*v1.Host)

	var jobRef string
	if job != nil {
		jobRef = job.Metadata.Name
		if reason == "" {
			reason = job.Status.GetCondition(core.ConditionTypeRun)
		}
	}

	if err := o.recordEvent(Event{
		BaseApiObj: host.BaseApiObj,
		Action:     action,
		Msg:        reason,
		JobRef:     jobRef,
		Phase:      core.PhaseFailed,
	}); err != nil {
		log.Error(err)
	}

	switch action {
	case core.EventActionConnect:
		host.Status.SetCondition(core.ConditionTypeConnected, reason)
	case core.EventActionInitial:
		host.Status.SetCondition(core.ConditionTypeInitialized, reason)
	}

	host.SetStatusPhase(core.PhaseNotReady)
	if _, err := o.helper.V1.Host.UpdateStatus(host.Metadata.Namespace, host.Metadata.Name, host.Status); err != nil {
		log.Error(err)
	}
}

// NewHostOperator 创建主机管理器
func NewHostOperator() *HostOperator {
	o := &HostOperator{
		BaseOperator: NewBaseOperator(v1.NewHostRegistry()),
	}
	o.SetHandleFunc(o.handleHost)
	o.SetFinalizeFunc(o.finalizeHost)
	o.SetReconcileFunc(o.reconcileHost, 30)
	return o
}
