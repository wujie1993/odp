package dpctl

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"

	clientset "github.com/wujie1993/waves/pkg/client"
	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v2"
)

type HostClient struct {
	clientset.ClientSet
}

func (c HostClient) GetPrint(namespace string, name string, format string) error {
	if name != "" {
		host, err := c.ClientSet.V2().Hosts().Get(context.TODO(), name)
		if err != nil {
			log.Error(err)
			return err
		}
		return printHosts([]v2.Host{*host}, format)
	}

	hosts, err := c.ClientSet.V2().Hosts().List(context.TODO())
	if err != nil {
		log.Error(err)
		return err
	}
	return printHosts(hosts, format)
}

func (c HostClient) Apply(obj core.ApiObject) (core.ApiObject, error) {
	getObj, err := c.ClientSet.V2().Hosts().Get(context.TODO(), obj.GetMetadata().Name)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		log.Error(err)
		return nil, err
	}
	if getObj == nil {
		// 创建主机
		return c.Create(obj)
	}

	// 更新主机
	return c.Update(obj)
}

func (c HostClient) Create(obj core.ApiObject) (core.ApiObject, error) {
	// 转换成最新v2版本结构
	obj, err := orm.Convert(obj, core.GVK{Group: core.Group, ApiVersion: v2.ApiVersion, Kind: core.KindHost})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	host := obj.(*v2.Host)

	result, err := c.ClientSet.V2().Hosts().Create(context.TODO(), host)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	fmt.Printf("%s created\n", result.GetKey())

	return result, nil
}

func (c HostClient) Update(obj core.ApiObject) (core.ApiObject, error) {
	// 转换成最新v2版本结构
	obj, err := orm.Convert(obj, core.GVK{Group: core.Group, ApiVersion: v2.ApiVersion, Kind: core.KindHost})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	host := obj.(*v2.Host)

	result, err := c.ClientSet.V2().Hosts().Update(context.TODO(), host)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	fmt.Printf("%s updated\n", result.GetKey())

	return result, nil
}

func (c HostClient) Delete(namespace string, name string) (core.ApiObject, error) {
	return c.ClientSet.V2().Hosts().Delete(context.TODO(), name)
}

func printHosts(hosts []v2.Host, format string) error {
	switch format {
	case OutputFormatJSON:
		data, err := ToJSON(hosts, false)
		if err != nil {
			log.Error(err)
			return err
		}
		fmt.Println(string(data))
	case OutputFormatJSONPretty:
		data, err := ToJSON(hosts, true)
		if err != nil {
			log.Error(err)
			return err
		}
		fmt.Println(string(data))
	case OutputFormatYAML:
		for _, host := range hosts {
			data, err := host.ToYAML()
			if err != nil {
				log.Error(err)
				return err
			}
			fmt.Println("---")
			fmt.Print(string(data))
		}
	default:
		table := tablewriter.NewWriter(os.Stdout)
		table.SetBorder(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeader([]string{"name", "alias", "status", "gpus"})

		for _, host := range hosts {
			var desc string
			if shortName, ok := host.Metadata.Annotations["ShortName"]; ok && shortName != "" {
				desc = shortName
			}

			table.Append([]string{
				host.Metadata.Name,
				desc,
				host.Status.Phase,
				fmt.Sprint(len(host.Info.GPUs)),
			})
		}
		table.Render()
	}
	return nil
}
