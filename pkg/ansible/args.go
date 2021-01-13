package ansible

import (
	"strings"
)

var internalArgs []string

func init() {
	// 以自定义参数方式传递进来的特殊参数，会变为内置参数而非模块参数
	internalArgs = []string{
		"algorithm_plugin_name",
		"algorithm_plugin_version",
		"algorithm_gpu_ids",
		"algorithm_media_type",
		"algorithm_request_gpu",
		"deploy_dir",
		"data_prefix",
		"logs_prefix",
		"purge_data",
		"enable_logging",
	}
}

type AppArgs map[string]interface{}

func (args AppArgs) Set(key string, value interface{}) {
	if in(strings.ToLower(key), internalArgs) {
		args[strings.ToLower(key)] = value
	} else {
		moduleArgs := make(map[string]interface{})
		moduleArgsObj, ok := args["module_args"]
		if !ok {
			args["module_args"] = moduleArgs
		} else {
			moduleArgs = moduleArgsObj.(map[string]interface{})
		}
		moduleArgs[key] = value
	}
}

func NewAppArgs() AppArgs {
	return make(map[string]interface{})
}

func in(key string, list []string) bool {
	for _, item := range list {
		if key == item {
			return true
		}
	}
	return false
}
