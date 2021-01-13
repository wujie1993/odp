package dpctl

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"

	clientset "github.com/wujie1993/waves/pkg/client"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

type ConfigMapClient struct {
	clientset.ClientSet
}

func (c ConfigMapClient) GetPrint(namespace string, name string, format string) error {
	if name != "" {
		configMap, err := c.ClientSet.V1().ConfigMaps(namespace).Get(context.TODO(), name)
		if err != nil {
			log.Error(err)
			return err
		}
		return printConfigMaps([]v1.ConfigMap{*configMap}, format)
	}

	configMaps, err := c.ClientSet.V1().ConfigMaps(namespace).List(context.TODO())
	if err != nil {
		log.Error(err)
		return err
	}
	return printConfigMaps(configMaps, format)
}

func (c ConfigMapClient) Apply(obj core.ApiObject) (core.ApiObject, error) {
	meta := obj.GetMetadata()

	getObj, err := c.ClientSet.V1().ConfigMaps(meta.Namespace).Get(context.TODO(), meta.Name)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		log.Error(err)
		return nil, err
	}
	if getObj == nil {
		return c.Create(obj)
	}

	return c.Update(obj)
}

func (c ConfigMapClient) Create(obj core.ApiObject) (core.ApiObject, error) {
	configMap := obj.(*v1.ConfigMap)

	result, err := c.ClientSet.V1().ConfigMaps(configMap.Metadata.Namespace).Create(context.TODO(), configMap)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	fmt.Printf("%s created\n", result.GetKey())

	return result, nil
}

func (c ConfigMapClient) Update(obj core.ApiObject) (core.ApiObject, error) {
	configMap := obj.(*v1.ConfigMap)

	result, err := c.ClientSet.V1().ConfigMaps(configMap.Metadata.Namespace).Update(context.TODO(), configMap)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	fmt.Printf("%s updated\n", result.GetKey())

	return result, nil
}

func (c ConfigMapClient) Delete(namespace string, name string) (core.ApiObject, error) {
	return c.ClientSet.V1().ConfigMaps(namespace).Delete(context.TODO(), name)
}

func printConfigMaps(configMaps []v1.ConfigMap, format string) error {
	switch format {
	case OutputFormatJSON:
		data, err := ToJSON(configMaps, false)
		if err != nil {
			log.Error(err)
			return err
		}
		fmt.Println(string(data))
	case OutputFormatJSONPretty:
		data, err := ToJSON(configMaps, true)
		if err != nil {
			log.Error(err)
			return err
		}
		fmt.Println(string(data))
	case OutputFormatYAML:
		for _, configMap := range configMaps {
			data, err := configMap.ToYAML()
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
		table.SetHeader([]string{"namespace", "name", "last updates", "revision"})
		for _, configMap := range configMaps {
			table.Append([]string{
				configMap.Metadata.Namespace,
				configMap.Metadata.Name,
				configMap.Metadata.UpdateTime.Format("2006/1/2 15:04:05"),
				fmt.Sprint(configMap.Metadata.ResourceVersion),
			})
		}
		table.Render()
	}
	return nil
}
