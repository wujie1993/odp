package schedule

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/ansible"
	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/file"
	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/orm/v2"
	"github.com/wujie1993/waves/pkg/setting"
)

type Worker struct {
	busy  bool
	mutex sync.Mutex
	Hash  string
}

func (w Worker) IsBusy() bool {
	return w.busy
}

func (w *Worker) Run(ctx context.Context, job *v2.Job) error {
	w.mutex.Lock()
	w.busy = true
	defer func() {
		w.mutex.Unlock()
		w.busy = false
	}()

	jobDir, _ := filepath.Abs(filepath.Join(setting.AppSetting.DataDir, setting.JobsDir, job.Metadata.Uid))

	// 清理任务目录
	os.RemoveAll(jobDir)

	// 创建任务目录
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		log.Error(err)
		return err
	}

	helper := orm.GetHelper()

	cmds := []ansible.RunCMD{}

	for playIndex, play := range job.Spec.Exec.Ansible.Plays {
		playDir := filepath.Join(jobDir, fmt.Sprintf("%d-%s", playIndex, play.Name))

		// 创建playbook工作目录
		if err := os.MkdirAll(playDir, 0755); err != nil {
			log.Error(err)
			return err
		}

		cmd := []string{job.Spec.Exec.Ansible.Bin}

		// 生成group_vars文件
		groupVarsFilename := "group_vars.yml"
		groupVarsPath := filepath.Join(playDir, groupVarsFilename)
		if play.GroupVars.ValueFrom.ConfigMapRef.Namespace != "" && play.GroupVars.ValueFrom.ConfigMapRef.Name != "" {
			obj, err := helper.V1.ConfigMap.Get(context.TODO(), play.GroupVars.ValueFrom.ConfigMapRef.Namespace, play.GroupVars.ValueFrom.ConfigMapRef.Name)
			if err != nil {
				log.Error(err)
				return err
			} else if obj == nil {
				err := e.Errorf("configMap %s/%s not found", play.GroupVars.ValueFrom.ConfigMapRef.Namespace, play.GroupVars.ValueFrom.ConfigMapRef.Name)
				log.Error(err)
				return err
			}
			cm := obj.(*v1.ConfigMap)

			if play.GroupVars.ValueFrom.ConfigMapRef.Key != "" {
				dataValue, ok := cm.Data[play.GroupVars.ValueFrom.ConfigMapRef.Key]
				if !ok {
					err := e.Errorf("configMap %s/%s not contain with key %s", play.GroupVars.ValueFrom.ConfigMapRef.Namespace, play.GroupVars.ValueFrom.ConfigMapRef.Name, play.GroupVars.ValueFrom.ConfigMapRef.Key)
					log.Error(err)
					return err
				}
				// 写入配置文件
				if err := file.Append(groupVarsPath, []byte(dataValue), 0755); err != nil {
					log.Error(err)
					return err
				}
			} else {
				for _, dataValue := range cm.Data {
					// 写入配置文件
					if err := file.Append(groupVarsPath, []byte(dataValue), 0755); err != nil {
						log.Error(err)
						return err
					}
				}
			}
		} else {
			// 写入配置文件
			if err := file.Append(groupVarsPath, []byte(play.GroupVars.Value), 0755); err != nil {
				log.Error(err)
				return err
			}
		}

		// 生成inventory文件
		inventoryFilename := "inventory.yml"
		inventoryPath := filepath.Join(playDir, inventoryFilename)
		if play.Inventory.ValueFrom.ConfigMapRef.Namespace != "" && play.Inventory.ValueFrom.ConfigMapRef.Name != "" {
			obj, err := helper.V1.ConfigMap.Get(context.TODO(), play.Inventory.ValueFrom.ConfigMapRef.Namespace, play.Inventory.ValueFrom.ConfigMapRef.Name)
			if err != nil {
				log.Error(err)
				return err
			}
			if obj == nil {
				err := e.Errorf("configmap %s/%s not found", play.Inventory.ValueFrom.ConfigMapRef.Namespace, play.Inventory.ValueFrom.ConfigMapRef.Name)
				log.Error(err)
				return err
			}
			cm := obj.(*v1.ConfigMap)

			if play.Inventory.ValueFrom.ConfigMapRef.Key != "" {
				inventoryData, ok := cm.Data[play.Inventory.ValueFrom.ConfigMapRef.Key]
				if !ok {
					err := e.Errorf("configMap %s/%s not contain with key %s", play.Inventory.ValueFrom.ConfigMapRef.Namespace, play.Inventory.ValueFrom.ConfigMapRef.Name, play.Inventory.ValueFrom.ConfigMapRef.Key)
					log.Error(err)
					return err
				}
				if err := file.Append(inventoryPath, []byte(inventoryData), 0755); err != nil {
					log.Error(err)
					return err
				}
				cmd = append(cmd, "-i", inventoryPath)
			} else {
				for _, inventoryData := range cm.Data {
					if err := file.Append(inventoryPath, []byte(inventoryData), 0755); err != nil {
						log.Error(err)
						return err
					}
					cmd = append(cmd, "-i", inventoryPath)
				}
			}
		} else {
			if err := file.Append(inventoryPath, []byte(play.Inventory.Value), 0755); err != nil {
				log.Error(err)
				return err
			}
			cmd = append(cmd, "-i", inventoryPath)
		}

		// 生成环境参数
		for _, env := range play.Envs {
			cmd = append(cmd, "-e", env)
		}

		// 生成标签参数
		if len(play.Tags) > 0 {
			cmd = append(cmd, "--tags", strings.Join(play.Tags, ","))
		}

		// 生成playbook文件
		playbookFilename := "playbook.yml"
		playbookPath := filepath.Join(playDir, playbookFilename)
		if play.Playbook.ValueFrom.ConfigMapRef.Namespace != "" && play.Playbook.ValueFrom.ConfigMapRef.Name != "" {
			obj, err := helper.V1.ConfigMap.Get(context.TODO(), play.Playbook.ValueFrom.ConfigMapRef.Namespace, play.Playbook.ValueFrom.ConfigMapRef.Name)
			if err != nil {
				log.Error(err)
				return err
			}
			if obj == nil {
				err := e.Errorf("configmap %s/%s not found", play.Playbook.ValueFrom.ConfigMapRef.Namespace, play.Playbook.ValueFrom.ConfigMapRef.Name)
				log.Error(err)
				return err
			}
			cm := obj.(*v1.ConfigMap)

			if play.Playbook.ValueFrom.ConfigMapRef.Key != "" {
				playbookData, ok := cm.Data[play.Playbook.ValueFrom.ConfigMapRef.Key]
				if !ok {
					err := e.Errorf("configMap %s/%s not contain with key %s", play.Playbook.ValueFrom.ConfigMapRef.Namespace, play.Playbook.ValueFrom.ConfigMapRef.Name, play.Playbook.ValueFrom.ConfigMapRef.Key)
					log.Error(err)
					return err
				}
				if err := file.Append(playbookPath, []byte(playbookData), 0755); err != nil {
					log.Error(err)
					return err
				}
				cmd = append(cmd, playbookPath)
			} else {
				for _, playbookData := range cm.Data {
					if err := file.Append(playbookPath, []byte(playbookData), 0755); err != nil {
						log.Error(err)
						return err
					}
					cmd = append(cmd, playbookPath)
				}
			}
		} else {
			if err := file.Append(playbookPath, []byte(play.Playbook.Value), 0755); err != nil {
				log.Error(err)
				return err
			}
			cmd = append(cmd, playbookPath)
		}

		// 生成configs文件
		for _, config := range play.Configs {
			if config.ValueFrom.ConfigMapRef.Namespace == "" || config.ValueFrom.ConfigMapRef.Name == "" {
				continue
			}

			obj, err := helper.V1.ConfigMap.Get(context.TODO(), config.ValueFrom.ConfigMapRef.Namespace, config.ValueFrom.ConfigMapRef.Name)
			if err != nil {
				log.Error(err)
				return err
			} else if obj == nil {
				err := errors.New(fmt.Sprintf("configMap %s/%s not found", config.ValueFrom.ConfigMapRef.Namespace, config.ValueFrom.ConfigMapRef.Name))
				log.Error(err)
				return err
			}

			cm := obj.(*v1.ConfigMap)
			for dataKey, dataValue := range cm.Data {
				path := filepath.Join(playDir, config.PathPrefix, dataKey)
				// 创建配置文件目录
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					log.Error(err)
					return err
				}
				// 写入配置文件
				if err := ioutil.WriteFile(path, []byte(dataValue), 0755); err != nil {
					log.Error(err)
					return err
				}
			}
		}

		// 将playbook运行命令写入run.sh
		cmds = append(cmds, ansible.RunCMD{
			Command:  strings.Join(cmd, " "),
			Reckless: job.Spec.Exec.Ansible.RecklessMode,
		})
	}

	runFilename := filepath.Join(jobDir, "run.sh")
	runFile, err := os.OpenFile(runFilename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Error(err)
		return err
	}
	defer runFile.Close()
	runSh, err := ansible.RenderRunShell(cmds)
	if err != nil {
		log.Error(err)
		return err
	}
	if _, err := runFile.WriteString(runSh); err != nil {
		log.Error(err)
		return err
	}
	runFile.Close()

	// 生成ansible.cfg
	cfgFilename := filepath.Join(jobDir, "ansible.cfg")
	cfgFile, err := os.OpenFile(cfgFilename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Error(err)
		return err
	}
	defer cfgFile.Close()
	cfgTpl, err := template.New("ansible_cfg.tpl").ParseFiles(filepath.Join(setting.AnsibleSetting.TplsDir, "ansible_cfg.tpl"))
	if err != nil {
		log.Error(err)
		return err
	}
	rolesDir, _ := filepath.Abs(setting.AnsibleSetting.PlaybooksDir)
	if err := cfgTpl.Execute(cfgFile, rolesDir); err != nil {
		log.Error(err)
		return err
	}

	// 执行run.sh
	var cmd *exec.Cmd
	if setting.AnsibleSetting.DryRun {
		cmd = exec.CommandContext(ctx, "/usr/bin/echo", "dry run with "+runFilename)
	} else {
		cmd = exec.CommandContext(ctx, "/usr/bin/sh", runFilename)
	}
	cmd.Dir = jobDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Error(err)
		return err
	}

	errReader, err := cmd.StderrPipe()
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debug(cmd.String())

	startTime := time.Now()

	if err := cmd.Start(); err != nil {
		log.Error(err)
		return err
	}

	// create log file in job execute dir
	logFile, err := os.OpenFile(filepath.Join(jobDir, "ansible.log"), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Error(err)
		return err
	}
	defer logFile.Close()

	reader := bufio.NewReader(stdout)
	logFile.WriteString("WORKDIR " + jobDir + "\n")
	logFile.WriteString("CMD " + cmd.String() + "\n")
	// Read and print standard output
	for {
		data, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error(err)
			cmd.Wait()
			return err
		}
		logFile.Write(data)
		logFile.Write([]byte{'\n'})
		if setting.AnsibleSetting.LogToStdout {
			fmt.Println(string(data))
		}
	}

	io.Copy(logFile, errReader)

	if err := cmd.Wait(); err != nil {
		log.Error(err)

		elapsed := time.Since(startTime)
		if elapsed >= job.Spec.TimeoutSeconds*time.Second || err.Error() == context.DeadlineExceeded.Error() {
			return e.Errorf("任务执行超时")
		}
		return err
	}
	return nil
}
