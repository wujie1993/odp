package operators

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wujie1993/waves/pkg/ansible"
	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/setting"
)

type Params struct {
	Groups []Group
}

type Group struct {
	Hosts []Host
	Name  string
}

type Host struct {
	Addr string
	User string
	Pass string
	Port uint16
}

type AllHosts struct {
	AllHosts []Label
}
type Label struct {
	Host     string
	User     string
	Password string
	Port     uint16
	Vars     map[string]string
}

type K8sInstallOperator struct {
	BaseOperator
	K8sInstallHashs map[string]string
}

//安装k8s
func (c *K8sInstallOperator) handleK8s(ctx context.Context, obj core.ApiObject) error {
	k8s := obj.(*v1.K8sConfig)
	helper := c.helper
	var params Params

	// 构造结构体

	if len(k8s.Spec.K8SWorker.Hosts) != 0 {
		group := Group{Name: "k8s-worker"}
		hosts := []Host{}
		for _, v := range k8s.Spec.K8SWorker.Hosts {
			result, err := helper.V1.Host.Get(context.TODO(), "default", v.ValueFrom.HostRef)
			if err != nil {
				log.Error(err)
				return err
			}
			var host Label
			host.Host = result.(*v1.Host).Spec.SSH.Host
			host.User = result.(*v1.Host).Spec.SSH.User
			host.Password = result.(*v1.Host).Spec.SSH.Password
			host.Port = result.(*v1.Host).Spec.SSH.Port
			host.Vars = v.Label
			// 构建结构体
			hosts = append(hosts, Host{
				Addr: host.Host,
				User: host.User,
				Pass: host.Password,
				Port: host.Port,
			})
		}
		group.Hosts = hosts
		params.Groups = append(params.Groups, group)
	}
	if len(k8s.Spec.K8SMaster.Hosts) != 0 {
		master_result, err := helper.V1.Host.Get(context.TODO(), "default", k8s.Spec.K8SMaster.Hosts[0].ValueFrom.HostRef)
		if err != nil {
			log.Error(err)
			return err
		}
		name := []string{"etcd", "harbor", "k8s-master"}

		for _, v := range name {
			params.Groups = append(params.Groups, Group{
				Name: v,
				Hosts: []Host{
					Host{
						Addr: master_result.(*v1.Host).Spec.SSH.Host,
						User: master_result.(*v1.Host).Spec.SSH.User,
						Pass: master_result.(*v1.Host).Spec.SSH.Password,
						Port: master_result.(*v1.Host).Spec.SSH.Port,
					},
				},
			})
		}
	}
	if len(k8s.Spec.K8SWorkerNew.Hosts) != 0 {
		group := Group{Name: "k8s-worker-new"}
		hosts := []Host{}
		for _, v := range k8s.Spec.K8SWorkerNew.Hosts {
			result, err := helper.V1.Host.Get(context.TODO(), "default", v.ValueFrom.HostRef)
			if err != nil {
				log.Error(err)
				return err
			}
			var host Label
			host.Host = result.(*v1.Host).Spec.SSH.Host
			host.User = result.(*v1.Host).Spec.SSH.User
			host.Password = result.(*v1.Host).Spec.SSH.Password
			host.Port = result.(*v1.Host).Spec.SSH.Port
			host.Vars = v.Label
			// 构建结构体
			hosts = append(hosts, Host{
				Addr: host.Host,
				User: host.User,
				Pass: host.Password,
				Port: host.Port,
			})
		}
		group.Hosts = hosts
		params.Groups = append(params.Groups, group)
	}
	// inventory 构建
	inventoryTpl, err := template.New("inventory").Parse(ansible.K8s_INVENTORY_TPL)
	if err != nil {
		log.Error(err)
		return err
	}
	var inventoryBuf bytes.Buffer
	if err := inventoryTpl.Execute(&inventoryBuf, params); err != nil {
		log.Error(err)
	}
	// playbook 构建
	var k8sinventoryBuf bytes.Buffer
	var action string
	// 如果不是新增节点，就是默认playbook
	if len(k8s.Spec.K8SWorkerNew.Hosts) == 0 {
		// install_template, err := template.New("inventory").Parse(ansible.ANSIBLE_K8SINSTALL_TPL)
		action = core.EventActionInstall
		install_template, err := template.New("k8s_install.tpl").ParseFiles(filepath.Join(setting.AnsibleSetting.TplsDir, "k8s_install.tpl"))

		if err != nil {
			log.Print(err)
			return err
		}
		if err := install_template.Execute(&k8sinventoryBuf, nil); err != nil {
			log.Error(err)
		}

	} else {
		action = core.EventActionInstallNode
		install_template, err := template.New("k8s_newworker.tpl").ParseFiles(filepath.Join(setting.AnsibleSetting.TplsDir, "k8s_newworker.tpl"))
		if err != nil {
			log.Print(err)
			return err
		}
		if err := install_template.Execute(&k8sinventoryBuf, nil); err != nil {
			log.Error(err)
		}
	}

	switch k8s.Status.Phase {
	case core.PhaseWaiting:
		switch k8s.Spec.Action {
		case core.PhaseInstalled:

			commonInventoryStr, err := ansible.RenderCommonInventory()
			if err != nil {
				log.Error(err)
				return err
			}
			// log.Debug("yaml", &k8sinventoryBuf)
			// 创建job
			job := v1.NewJob()
			job.Metadata.Namespace = "default"
			job.Metadata.Name = fmt.Sprintf("%s-%s-%d", "k8sinstall", k8s.Metadata.Name, time.Now().Unix())
			job.Spec.Exec.Type = core.JobExecTypeAnsible
			job.Spec.Exec.Ansible.Bin = "/usr/bin/ansible-playbook"
			job.Spec.Exec.Ansible.Inventories = []v1.AnsibleInventory{
				v1.AnsibleInventory{Value: commonInventoryStr},
				v1.AnsibleInventory{Value: inventoryBuf.String()},
			}
			log.Debug("inventory", job.Spec.Exec.Ansible.Inventories)
			job.Spec.Exec.Ansible.Envs = []string{
				"act=install",
			}
			job.Spec.Exec.Ansible.Playbook = k8sinventoryBuf.String()
			job.Spec.TimeoutSeconds = 6000
			job.Spec.FailureThreshold = 1

			if _, err := helper.V1.Job.Create(context.TODO(), job); err != nil {
				log.Error(err)
				return err
			}

			// 记录事件开始
			if err := c.recordEvent(Event{
				BaseApiObj: k8s.BaseApiObj,
				Action:     action,
				Msg:        "",
				JobRef:     job.Metadata.Name,
				Phase:      core.PhaseWaiting,
			}); err != nil {
				log.Error(err)
			}

			// 监听任务
			jobCtx, _ := context.WithCancel(ctx)
			jobActionChan := c.helper.V1.Job.Watch(jobCtx, "", job.Metadata.Name)
			for jobAction := range jobActionChan {
				job := jobAction.Obj.(*v1.Job)
				switch jobAction.Type {
				case db.KVActionTypeDelete:
					return nil
				case db.KVActionTypeSet:
					switch job.Status.Phase {
					case core.PhaseCompleted:
						// k8s.Status.SetCondition(core.ConditionTypeInitialized, core.ConditionStatusTrue)
						if len(k8s.Spec.K8SWorkerNew.Hosts) == 0 {
							k8s.Metadata.Annotations["k8sinstall"] = core.PhaseCompleted
							if err, _ := c.registry.Update(context.TODO(), k8s); err != nil {
								log.Error(err)
							}
						}
						k8s.SetStatusPhase(core.PhaseCompleted)
						if _, err := c.helper.V1.K8sConfig.UpdateStatus(k8s.Metadata.Namespace, k8s.Metadata.Name, k8s.Status); err != nil {
							log.Error(err)
						}

						// 记录事件完成
						if err := c.recordEvent(Event{
							BaseApiObj: k8s.BaseApiObj,
							Action:     action,
							Msg:        "",
							JobRef:     job.Metadata.Name,
							Phase:      core.PhaseCompleted,
						}); err != nil {
							log.Error(err)
						}

						//打标签
						c.handleK8sLabel(ctx, k8s)
						return nil
					case core.PhaseFailed:
						if len(k8s.Spec.K8SWorkerNew.Hosts) == 0 {
							k8s.Metadata.Annotations["k8sinstall"] = core.PhaseFailed
							if err, _ := c.registry.Update(context.TODO(), k8s); err != nil {
								log.Error(err)
							}
						}
						k8s.SetStatusPhase(core.PhaseFailed)
						if _, err := c.helper.V1.K8sConfig.UpdateStatus(k8s.Metadata.Namespace, k8s.Metadata.Name, k8s.Status); err != nil {
							log.Error(err)
						}

						// 记录事件失败
						if err := c.recordEvent(Event{
							BaseApiObj: k8s.BaseApiObj,
							Action:     action,
							Msg:        "",
							JobRef:     job.Metadata.Name,
							Phase:      core.PhaseFailed,
						}); err != nil {
							log.Error(err)
						}
					case core.PhaseRunning:
						if len(k8s.Spec.K8SWorkerNew.Hosts) == 0 {
							k8s.Metadata.Annotations["k8sinstall"] = core.PhaseInstalling
							if err, _ := c.registry.Update(context.TODO(), k8s); err != nil {
								log.Error(err)
							}
						}
						k8s.SetStatusPhase(core.PhaseInstalling)
						if _, err := c.helper.V1.K8sConfig.UpdateStatus(k8s.Metadata.Namespace, k8s.Metadata.Name, k8s.Status); err != nil {
							log.Error(err)
						}

					default:
						log.Errorf("unknown status '%s' of job '%s'", job.Status.Phase, job.GetKey())
					}
				}
			}
			return nil
			//卸载k8s集群
		case core.PhaseUninstalled:
			c.deleteK8s(ctx, k8s, inventoryBuf)
		//卸载单个node节点
		case core.PhaseUninstallNode:
			c.deleteK8s(ctx, k8s, inventoryBuf)
		//已安装集群打标签
		case core.PhaseLabel:
			c.handleK8sLabel(ctx, k8s)
		// 删除标签功能
		case core.PhaseUnLabel:
			c.handleK8sLabel(ctx, k8s)

		}
	case core.PhaseDeleting:
		c.delete(ctx, obj)
	}

	return nil
}

func (o K8sInstallOperator) finalizeK8s(ctx context.Context, obj core.ApiObject) error {
	k8s := obj.(*v1.K8sConfig)

	// 每次只处理一项Finalizer
	switch k8s.Metadata.Finalizers[0] {
	case core.FinalizerCleanRefEvent:
		// 同步删除关联的事件
		eventList, err := o.helper.V1.Event.List(context.TODO(), "")
		if err != nil {
			log.Error(err)
			return err
		}
		for _, eventObj := range eventList {
			event := eventObj.(*v1.Event)
			if event.Spec.ResourceRef.Kind == core.KindK8sConfig && event.Spec.ResourceRef.Namespace == k8s.Metadata.Namespace && event.Spec.ResourceRef.Name == k8s.Metadata.Name {
				if _, err := o.helper.V1.Event.Delete(context.TODO(), "", event.Metadata.Name); err != nil {
					log.Error(err)
					return err
				}
			}
		}
	}
	return nil
}

//卸载k8s 集群
func (c *K8sInstallOperator) deleteK8s(ctx context.Context, k8s *v1.K8sConfig, inventoryBuf bytes.Buffer) {
	log.Debug("卸载k8s", k8s.Spec.K8SWorkerNew.Hosts)

	helper := c.helper
	//playbook 构建
	var k8suninstallinventoryBuf bytes.Buffer
	var action string
	if len(k8s.Spec.K8SMaster.Hosts) != 0 {
		action = core.EventActionUninstall
		install_template, err := template.New("k8s_uninstall.tpl").ParseFiles(filepath.Join(setting.AnsibleSetting.TplsDir, "k8s_uninstall.tpl"))
		if err != nil {
			log.Print(err)
			return
		}
		if err := install_template.Execute(&k8suninstallinventoryBuf, nil); err != nil {
			log.Error(err)
		}
	} else {
		action = core.EventActionUninstallNode
		install_template, err := template.New("k8s_uninstallnode.tpl").ParseFiles(filepath.Join(setting.AnsibleSetting.TplsDir, "k8s_uninstallnode.tpl"))
		if err != nil {
			log.Print(err)
			return
		}
		if err := install_template.Execute(&k8suninstallinventoryBuf, nil); err != nil {
			log.Error(err)
		}
	}
	commonInventoryStr, err := ansible.RenderCommonInventory()
	if err != nil {
		log.Error(err)
		return
	}

	job := v1.NewJob()
	job.Metadata.Namespace = "default"
	job.Metadata.Name = fmt.Sprintf("%s-%s-%d", "k8s", "uninstall", time.Now().Unix())
	job.Spec.Exec.Type = core.JobExecTypeAnsible
	job.Spec.Exec.Ansible.Bin = "/usr/bin/ansible-playbook"
	job.Spec.Exec.Ansible.Inventories = []v1.AnsibleInventory{
		v1.AnsibleInventory{Value: commonInventoryStr},
		v1.AnsibleInventory{Value: inventoryBuf.String()},
	}
	// job.Spec.Exec.Ansible.Envs = []string{
	// 	"act=install",
	// }
	job.Spec.Exec.Ansible.Playbook = k8suninstallinventoryBuf.String()
	job.Spec.TimeoutSeconds = 6000
	job.Spec.FailureThreshold = 1

	if _, err := helper.V1.Job.Create(context.TODO(), job); err != nil {
		log.Error(err)
		return
	}

	// 记录事件开始
	if err := c.recordEvent(Event{
		BaseApiObj: k8s.BaseApiObj,
		Action:     action,
		Msg:        "",
		JobRef:     job.Metadata.Name,
		Phase:      core.PhaseWaiting,
	}); err != nil {
		log.Error(err)
	}

	// 监听任务
	jobCtx, _ := context.WithCancel(ctx)
	jobActionChan := c.helper.V1.Job.Watch(jobCtx, "", job.Metadata.Name)
	for jobAction := range jobActionChan {
		job := jobAction.Obj.(*v1.Job)
		switch jobAction.Type {
		case db.KVActionTypeDelete:
			return
		case db.KVActionTypeSet:
			switch job.Status.Phase {
			case core.PhaseCompleted:
				if len(k8s.Spec.K8SMaster.Hosts) != 0 {
					k8s.Metadata.Annotations["k8sinstall"] = ""
					if err, _ := c.registry.Update(context.TODO(), k8s); err != nil {
						log.Error(err)
					}
				}
				k8s.Status.SetCondition(core.ConditionTypeInitialized, core.ConditionStatusTrue)
				k8s.SetStatusPhase(core.PhaseCompleted)
				if _, err := c.helper.V1.K8sConfig.UpdateStatus(k8s.Metadata.Namespace, k8s.Metadata.Name, k8s.Status); err != nil {
					log.Error(err)
				}

				// 记录事件完成
				if err := c.recordEvent(Event{
					BaseApiObj: k8s.BaseApiObj,
					Action:     action,
					Msg:        "",
					JobRef:     job.Metadata.Name,
					Phase:      core.PhaseCompleted,
				}); err != nil {
					log.Error(err)
				}
				return
			case core.PhaseFailed:
				if len(k8s.Spec.K8SMaster.Hosts) != 0 {
					k8s.Metadata.Annotations["k8sinstall"] = ""
					if err, _ := c.registry.Update(context.TODO(), k8s); err != nil {
						log.Error(err)
					}
				}
				k8s.Status.SetCondition(core.ConditionTypeInitialized, core.ConditionStatusFalse)
				k8s.SetStatusPhase(core.PhaseFailed)
				if _, err := c.helper.V1.K8sConfig.UpdateStatus(k8s.Metadata.Namespace, k8s.Metadata.Name, k8s.Status); err != nil {
					log.Error(err)
				}

				// 记录事件失败
				if err := c.recordEvent(Event{
					BaseApiObj: k8s.BaseApiObj,
					Action:     action,
					Msg:        "",
					JobRef:     job.Metadata.Name,
					Phase:      core.PhaseFailed,
				}); err != nil {
					log.Error(err)
				}

			case core.PhaseRunning:
				if len(k8s.Spec.K8SMaster.Hosts) != 0 {
					k8s.Metadata.Annotations["k8sinstall"] = ""
					if err, _ := c.registry.Update(context.TODO(), k8s); err != nil {
						log.Error(err)
					}
				}
				k8s.Status.SetCondition(core.ConditionTypeInitialized, core.ConditionStatusFalse)
				k8s.SetStatusPhase(core.PhaseUninstalling)
				if _, err := c.helper.V1.K8sConfig.UpdateStatus(k8s.Metadata.Namespace, k8s.Metadata.Name, k8s.Status); err != nil {
					log.Error(err)
				}

			default:
				log.Errorf("unknown status '%s' of job '%s'", job.Status.Phase, job.GetKey())
			}
		}
	}
	return
}

// 打标签
func (c *K8sInstallOperator) handleK8sLabel(ctx context.Context, k8s *v1.K8sConfig) {
	helper := c.helper

	var all_host AllHosts
	count := 0
	if len(k8s.Spec.K8SWorker.Hosts) != 0 {
		for _, v := range k8s.Spec.K8SWorker.Hosts {
			result, err := helper.V1.Host.Get(context.TODO(), "default", v.ValueFrom.HostRef)
			if err != nil {
				log.Error(err)
			}
			count += len(v.Label)

			var host Label
			host.Host = result.(*v1.Host).Spec.SSH.Host
			host.User = result.(*v1.Host).Spec.SSH.User
			host.Password = result.(*v1.Host).Spec.SSH.Password
			host.Port = result.(*v1.Host).Spec.SSH.Port
			host.Vars = v.Label
			all_host.AllHosts = append(all_host.AllHosts, Label{
				Host:     host.Host,
				User:     host.User,
				Password: host.Password,
				Port:     host.Port,
				Vars:     host.Vars,
			})
		}
	}
	if len(k8s.Spec.K8SWorkerNew.Hosts) != 0 {
		for _, v := range k8s.Spec.K8SWorkerNew.Hosts {
			result, err := helper.V1.Host.Get(context.TODO(), "default", v.ValueFrom.HostRef)
			if err != nil {
				log.Error(err)
			}
			count += len(v.Label)

			var host Label
			host.Host = result.(*v1.Host).Spec.SSH.Host
			host.User = result.(*v1.Host).Spec.SSH.User
			host.Password = result.(*v1.Host).Spec.SSH.Password
			host.Port = result.(*v1.Host).Spec.SSH.Port
			host.Vars = v.Label
			all_host.AllHosts = append(all_host.AllHosts, Label{
				Host:     host.Host,
				User:     host.User,
				Password: host.Password,
				Port:     host.Port,
				Vars:     host.Vars,
			})
		}
	}
	if len(k8s.Spec.K8SMaster.Hosts) != 0 {
		master_result, err := helper.V1.Host.Get(context.TODO(), "default", k8s.Spec.K8SMaster.Hosts[0].ValueFrom.HostRef)
		if err != nil {
			log.Error(err)
		}
		count += len(k8s.Spec.K8SMaster.Hosts[0].Label)

		all_host.AllHosts = append(all_host.AllHosts, Label{
			Host:     master_result.(*v1.Host).Spec.SSH.Host,
			User:     master_result.(*v1.Host).Spec.SSH.User,
			Password: master_result.(*v1.Host).Spec.SSH.Password,
			Port:     master_result.(*v1.Host).Spec.SSH.Port,
			Vars:     k8s.Spec.K8SMaster.Hosts[0].Label,
		})
	}

	log.Debug("标签数", count)
	if count == 0 {
		return
	}

	// 处于等待
	inventoryTpl, err := template.New("inventory").Parse(ansible.ANSIBLE_INVENTORY_LABEL_INIT_TPL)
	if err != nil {
		log.Error(err)
	}
	var inventoryBuf bytes.Buffer

	if err := inventoryTpl.Execute(&inventoryBuf, all_host.AllHosts); err != nil {
		log.Error(err)
	}
	log.Debug(&inventoryBuf)

	//playbooks
	install_template, err := template.New("inventory").Parse(ansible.NODE_LABEL_TPL)
	if err != nil {
		log.Print(err)
		return
	}
	var nodeinventoryBuf bytes.Buffer
	if err := install_template.Execute(&nodeinventoryBuf, nil); err != nil {
		log.Error(err)
		return
	}
	commonInventoryStr, err := ansible.RenderCommonInventory()
	if err != nil {
		log.Error(err)
		return
	}

	job := v1.NewJob()
	job.Metadata.Namespace = "default"
	job.Metadata.Name = fmt.Sprintf("%s-%s-%d", "k8slabel", k8s.Metadata.Name, time.Now().Unix())
	job.Spec.Exec.Type = core.JobExecTypeAnsible
	job.Spec.Exec.Ansible.Bin = "/usr/bin/ansible-playbook"
	job.Spec.Exec.Ansible.Inventories = []v1.AnsibleInventory{
		v1.AnsibleInventory{Value: commonInventoryStr},
		v1.AnsibleInventory{Value: inventoryBuf.String()},
	}
	//判断删除还是新增标签
	var action string
	if k8s.Spec.Action == core.PhaseLabel {
		action = core.EventActionLabel
		job.Spec.Exec.Ansible.Envs = []string{"act=install"}
	} else {
		action = core.EventActionUnLabel
		job.Spec.Exec.Ansible.Envs = []string{"act=uninstall"}
	}

	// job.Spec.Exec.Ansible.Envs = []string{
	// 	"act=configure",
	// }
	job.Spec.Exec.Ansible.Playbook = nodeinventoryBuf.String()
	job.Spec.TimeoutSeconds = 600
	job.Spec.FailureThreshold = 1

	if _, err := helper.V1.Job.Create(context.TODO(), job); err != nil {
		log.Error(err)
		return
	}

	// 记录事件开始
	if err := c.recordEvent(Event{
		BaseApiObj: k8s.BaseApiObj,
		Action:     action,
		Msg:        "",
		JobRef:     job.Metadata.Name,
		Phase:      core.PhaseWaiting,
	}); err != nil {
		log.Error(err)
	}

	// 监听任务
	jobCtx, _ := context.WithCancel(ctx)
	jobActionChan := c.helper.V1.Job.Watch(jobCtx, "", job.Metadata.Name)
	for jobAction := range jobActionChan {
		job := jobAction.Obj.(*v1.Job)
		switch jobAction.Type {
		case db.KVActionTypeDelete:
			return
		case db.KVActionTypeSet:
			switch job.Status.Phase {
			case core.PhaseCompleted:
				k8s.Status.SetCondition(core.ConditionTypeInitialized, core.ConditionStatusTrue)
				k8s.SetStatusPhase(core.PhaseCompleted)
				if _, err := c.helper.V1.K8sConfig.UpdateStatus(k8s.Metadata.Namespace, k8s.Metadata.Name, k8s.Status); err != nil {
					log.Error(err)
				}

				// 记录事件完成
				if err := c.recordEvent(Event{
					BaseApiObj: k8s.BaseApiObj,
					Action:     action,
					Msg:        "",
					JobRef:     job.Metadata.Name,
					Phase:      core.PhaseCompleted,
				}); err != nil {
					log.Error(err)
				}
				return
			case core.PhaseFailed:
				k8s.Status.SetCondition(core.ConditionTypeInitialized, core.ConditionStatusFalse)
				k8s.SetStatusPhase(core.PhaseFailed)
				if _, err := c.helper.V1.K8sConfig.UpdateStatus(k8s.Metadata.Namespace, k8s.Metadata.Name, k8s.Status); err != nil {
					log.Error(err)
				}

				// 记录事件失败
				if err := c.recordEvent(Event{
					BaseApiObj: k8s.BaseApiObj,
					Action:     action,
					Msg:        "",
					JobRef:     job.Metadata.Name,
					Phase:      core.PhaseFailed,
				}); err != nil {
					log.Error(err)
				}
			case core.PhaseRunning:
			default:
				log.Errorf("unknown status '%s' of job '%s'", job.Status.Phase, job.GetKey())
			}
		}
	}
	return
}

// NewK8sInstallOperator 创建K8S集群管理器
func NewK8sInstallOperator() *K8sInstallOperator {
	o := &K8sInstallOperator{
		BaseOperator: NewBaseOperator(v1.NewK8sConfigRegistry()),
	}
	o.SetHandleFunc(o.handleK8s)
	o.SetFinalizeFunc(o.finalizeK8s)
	return o
}
