package ansible

import (
	"bytes"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v2"

	"github.com/wujie1993/waves/pkg/setting"
)

const (
	ANSIBLE_GROUP_K8S_MASTER = "k8s-master"
	ConfigsDir               = "configs"
	AdditionalConfigsDir     = "upload_configs"
)

const (
	ANSIBLE_INVENTORY_HOST_INIT_TPL = `
centos-init:
  hosts:
  {{ range . }}
    {{ .Spec.SSH.Host }}:
      ansible_ssh_pass: {{ .Spec.SSH.Password }}
      ansible_ssh_port: {{ .Spec.SSH.Port }}
      ansible_ssh_user: {{ .Spec.SSH.User }}
  {{ end }}
`

	ANSIBLE_PLAYBOOK_TPL = `
{{ range . }}
- hosts:
  {{ range .Hosts }}
  - {{ . }}
  {{ end }}
  pre_tasks:
  {{ range .IncludeVars }}
  - include_vars: {{ . }}
    tags: always
  {{ end }}
  roles:
  {{ range .Roles }}
  - {{ . }}
  {{ end }}
{{ end }}
`

	ANSIBLE_CFG_TPL = `
[defaults]
roles_path = {{ . }}
host_key_checking = false
strategy = mitogen_linear
strategy_plugins = /usr/lib/python2.7/site-packages/ansible_mitogen/plugins/strategy
callback_whitelist = profile_tasks, dense
`

	RUN_SHELL_TPL = `
#!/bin/sh
rc=0
{{ range . }}
{{ .Command }}
if [ $? -ne 0 ]; then
    if {{ .Reckless }}; then
        rc=1
    else
        exit 1
    fi
fi
{{ end }}
exit $rc
`
)

const (
	ANSIBLE_ROLE_HOST_INIT = "centos-init"
)

type CommonVars struct {
	BaseDir string
}

type RunCMD struct {
	Command  string
	Reckless bool
}

type Inventory map[string]InventoryGroup

type InventoryGroup struct {
	Hosts map[string]InventoryHost `yaml:"hosts"`
	Vars  map[string]interface{}   `yaml:"vars"`
}

type InventoryHost map[string]interface{}

type Playbook struct {
	Hosts       []string
	Roles       []string
	IncludeVars []string
}

// RenderCommonInventory 渲染并返回公共Inventory
func RenderCommonInventory() (string, error) {
	basedir, _ := filepath.Abs(setting.AnsibleSetting.BaseDir)
	commonVars := CommonVars{
		BaseDir: basedir,
	}
	tpl, err := template.New("inventory_common.tpl").ParseFiles(filepath.Join(setting.AnsibleSetting.TplsDir, "inventory_common.tpl"))
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	if err := tpl.Execute(&buffer, commonVars); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

// RenderInventory 渲染并返回yaml格式的inventory字符串
func RenderInventory(inventory Inventory) (string, error) {
	data, err := yaml.Marshal(inventory)
	return string(data), err
}

// RenderRunShell 渲染并返回运行脚本字符串
func RenderRunShell(cmds []RunCMD) (string, error) {
	tpl, err := template.New("run").Parse(RUN_SHELL_TPL)
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	if err := tpl.Execute(&buffer, cmds); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

// RenderPlaybook 渲染并返回自定义Playbook
func RenderPlaybook(playbooks []Playbook) (string, error) {
	tpl, err := template.New("playbook").Parse(ANSIBLE_PLAYBOOK_TPL)
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	if err := tpl.Execute(&buffer, playbooks); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

const ANSIBLE_INVENTORY_LABEL_INIT_TPL = `
node_label_mgmt:
  hosts:
  {{ range . }}
    {{ .Host }}: 
      ansible_ssh_pass: {{ .Password }}
      ansible_ssh_port: {{ .Port }}
      ansible_ssh_user: {{ .User }}
    {{ range $key,$value := .Vars }}
      {{ $key}}: '{{$value}}'
    {{end}}
  {{end}}
`

var K8s_INVENTORY_TPL = `
{{ range .Groups }}
{{ .Name }}:
  hosts:
{{ range .Hosts }}
    {{ .Addr }}:
      ansible_ssh_pass: {{ .Pass }}
      ansible_ssh_port: {{ .Port }}
      ansible_ssh_user: {{ .User }}
  {{ end }}
{{ end }}
chrony: null
ex-lb: null
带GPU服务器: null
`

const NODE_LABEL_TPL = `
- hosts:
  - node_label_mgmt
  roles:
  - node_label_mgmt
`
