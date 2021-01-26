package wavectl

import (
	"fmt"
	"os"

	clientset "github.com/wujie1993/waves/pkg/client"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/wavectl/loader"
)

var exitCode int

func exit() {
	os.Exit(exitCode)
}

var clientSet clientset.ClientSet

type ResourceManager interface {
	GetPrint(namespace string, name string, format string) error
	Apply(obj core.ApiObject) (core.ApiObject, error)
	Create(obj core.ApiObject) (core.ApiObject, error)
	Update(obj core.ApiObject) (core.ApiObject, error)
	Delete(namespace string, name string) (core.ApiObject, error)
}

func initClient(endpoint string) {
	clientSet = clientset.NewClientSet(endpoint)
}

func getClient(kind string) ResourceManager {
	switch kind {
	case core.KindHost:
		return HostClient{
			ClientSet: clientSet,
		}
	case core.KindAppInstance:
		return AppInstanceClient{
			ClientSet: clientSet,
		}
	case core.KindConfigMap:
		return ConfigMapClient{
			ClientSet: clientSet,
		}
	case core.KindApp:
		return AppClient{
			ClientSet: clientSet,
		}
	}
	return nil
}

// GetResourceOptions 获取资源配置项
type GetResourceOptions struct {
	Endpoint     string
	Resource     string
	ResourceName string
	Namespace    string
	Format       string
}

// CreateResourceOptions 创建资源配置项
type CreateResourceOptions struct {
	Endpoint string
	File     string
}

// ApplyResourceOptions 资源应用配置项
type ApplyResourceOptions struct {
	Endpoint string
	File     string
}

// DeleteResourceOptions 资源删除配置项
type DeleteResourceOptions struct {
	Endpoint     string
	Namespace    string
	Resource     string
	ResourceName string
	File         string
}

// HostPluginOptions 主机插件操作配置项
type HostPluginOptions struct {
	Endpoint      string
	Action        string
	Host          string
	PluginName    string
	PluginVersion string
	Force         bool
}

// GetResource 获取资源
func GetResource(opts GetResourceOptions) {
	defer exit()

	kind := core.SearchKind(opts.Resource)

	initClient(opts.Endpoint)

	cli := getClient(kind)
	if cli == nil {
		fmt.Printf("does not support get %s\n", kind)
		exitCode++
		return
	}

	if err := cli.GetPrint(opts.Namespace, opts.ResourceName, opts.Format); err != nil {
		fmt.Println(err)
		exitCode++
		return
	}
}

// CreateResource 创建资源
func CreateResource(opts CreateResourceOptions) {
	defer exit()

	var objs []core.ApiObject

	if opts.File != "" {
		// 从本地文件加载对象
		fileObjs, err := loader.LoadObjsByLocalPath(opts.File)
		if err != nil {
			fmt.Println(err)
			exitCode++
			return
		}
		objs = fileObjs
	}

	initClient(opts.Endpoint)

	for _, obj := range objs {
		kind := obj.GetGVK().Kind

		cli := getClient(kind)
		if cli == nil {
			fmt.Printf("dose not support %s creation\n", kind)
			exitCode++
			continue
		}

		if _, err := cli.Create(obj); err != nil {
			fmt.Println(err)
			exitCode++
			continue
		}
	}
}

// ApplyResource 应用资源
func ApplyResource(opts ApplyResourceOptions) {
	defer exit()

	initClient(opts.Endpoint)

	/* 更新本地文件中指定的资源 */
	var objs []core.ApiObject

	// 从本地文件加载资源对象
	if opts.File != "" {
		fileObjs, err := loader.LoadObjsByLocalPath(opts.File)
		if err != nil {
			fmt.Println(err)
			exitCode++
			return
		}
		objs = fileObjs
	}

	// 更新资源对象
	for _, obj := range objs {
		kind := obj.GetGVK().Kind

		cli := getClient(kind)
		if cli == nil {
			fmt.Printf("dose not support %s creation\n", kind)
			exitCode++
			continue
		}

		if _, err := cli.Apply(obj); err != nil {
			fmt.Println(err)
			exitCode++
			continue
		}
	}
}

// DeleteResource 删除资源
func DeleteResource(opts DeleteResourceOptions) {
	defer exit()

	initClient(opts.Endpoint)

	/* 删除指定类型资源 */
	if opts.Resource != "" && opts.ResourceName != "" {
		kind := core.SearchKind(opts.Resource)
		cli := getClient(kind)
		if cli == nil {
			fmt.Printf("dose not support %s deletion\n", kind)
			exitCode++
			return
		}

		if _, err := cli.Delete(opts.Namespace, opts.ResourceName); err != nil {
			fmt.Println(err)
			exitCode++
		}
		return
	}

	/* 删除本地文件中指定的资源 */
	var objs []core.ApiObject

	// 从本地文件加载资源对象
	if opts.File != "" {
		fileObjs, err := loader.LoadObjsByLocalPath(opts.File)
		if err != nil {
			fmt.Println(err)
			exitCode++
			return
		}
		objs = fileObjs
	}

	// 删除资源对象
	for _, obj := range objs {
		kind := obj.GetGVK().Kind
		meta := obj.GetMetadata()

		cli := getClient(kind)
		if cli == nil {
			fmt.Printf("dose not support %s deletion\n", kind)
			exitCode++
			continue
		}

		if _, err := cli.Delete(meta.Namespace, meta.Name); err != nil {
			fmt.Println(err)
			exitCode++
			continue
		}
	}
}

func ManageHostPlugin(opts HostPluginOptions) {
	defer exit()

	initClient(opts.Endpoint)

	cli := HostPluginClient{
		ClientSet: clientSet,
	}
	switch opts.Action {
	case core.AppActionInstall:
		if err := cli.Install(opts.Host, opts.PluginName, opts.PluginVersion, opts.Force); err != nil {
			fmt.Println(err)
			exitCode++
			return
		}
	case core.AppActionUninstall:
		if err := cli.Uninstall(opts.Host, opts.PluginName, opts.Force); err != nil {
			fmt.Println(err)
			exitCode++
			return
		}
	default:
		fmt.Printf("unsupported action %s\n", opts.Action)
		exitCode++
		return
	}
}
