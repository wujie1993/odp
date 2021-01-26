package wavectl

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

type AppInstanceClient struct {
	clientset.ClientSet
}

func (c AppInstanceClient) GetPrint(namespace string, name string, format string) error {
	if name != "" {
		appInstance, err := c.ClientSet.V2().AppInstances(namespace).Get(context.TODO(), name)
		if err != nil {
			log.Error(err)
			return err
		}
		return printAppInstances([]v2.AppInstance{*appInstance}, format)
	}

	appInstances, err := c.ClientSet.V2().AppInstances(namespace).List(context.TODO())
	if err != nil {
		log.Error(err)
		return err
	}

	return printAppInstances(appInstances, format)
}

func (c AppInstanceClient) Apply(obj core.ApiObject) (core.ApiObject, error) {
	meta := obj.GetMetadata()

	getObj, err := c.ClientSet.V2().AppInstances(meta.Namespace).Get(context.TODO(), meta.Name)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		log.Error(err)
		return nil, err
	}
	if getObj == nil {
		return c.Create(obj)
	}

	return c.Update(obj)
}

func (c AppInstanceClient) Create(obj core.ApiObject) (core.ApiObject, error) {
	// 转换成v2版本结构
	obj, err := orm.Convert(obj, core.GVK{Group: core.Group, ApiVersion: v2.ApiVersion, Kind: core.KindAppInstance})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	appInstance := obj.(*v2.AppInstance)

	result, err := c.ClientSet.V2().AppInstances(appInstance.Metadata.Namespace).Create(context.TODO(), appInstance)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	fmt.Printf("%s created\n", result.GetKey())

	return result, nil
}

func (c AppInstanceClient) Update(obj core.ApiObject) (core.ApiObject, error) {
	// 转换成v2版本结构
	obj, err := orm.Convert(obj, core.GVK{Group: core.Group, ApiVersion: v2.ApiVersion, Kind: core.KindAppInstance})
	if err != nil {
		return nil, err
	}
	appInstance := obj.(*v2.AppInstance)

	result, err := c.ClientSet.V2().AppInstances(appInstance.Metadata.Namespace).Update(context.TODO(), appInstance)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	fmt.Printf("%s updated\n", result.GetKey())

	return result, nil
}

func (c AppInstanceClient) Delete(namespace string, name string) (core.ApiObject, error) {
	return c.ClientSet.V2().AppInstances(namespace).Delete(context.TODO(), name)
}

func printAppInstances(appInstances []v2.AppInstance, format string) error {
	switch format {
	case OutputFormatJSON:
		data, err := ToJSON(appInstances, false)
		if err != nil {
			log.Error(err)
			return err
		}
		fmt.Println(string(data))
	case OutputFormatJSONPretty:
		data, err := ToJSON(appInstances, true)
		if err != nil {
			log.Error(err)
			return err
		}
		fmt.Println(string(data))
	case OutputFormatYAML:
		for _, appInstance := range appInstances {
			data, err := appInstance.ToYAML()
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
		table.SetHeader([]string{"namespace", "name", "alias", "category", "app version", "status", "healthy", "last updates", "revision"})
		for _, appInstance := range appInstances {
			var desc string
			if shortName, ok := appInstance.Metadata.Annotations["ShortName"]; ok && shortName != "" {
				desc = shortName
			}

			healthy := appInstance.Status.GetCondition(core.ConditionTypeHealthy)
			if healthy != core.ConditionStatusTrue {
				healthy = core.ConditionStatusFalse
			}

			table.Append([]string{
				appInstance.Metadata.Namespace,
				appInstance.Metadata.Name,
				desc,
				core.GetCategoryMsg(appInstance.Spec.Category),
				appInstance.Spec.AppRef.Name + "-" + appInstance.Spec.AppRef.Version,
				appInstance.Status.Phase,
				healthy,
				appInstance.Metadata.UpdateTime.Format("2006/1/2 15:04:05"),
				fmt.Sprint(appInstance.Metadata.ResourceVersion),
			})
		}
		table.Render()
	}
	return nil
}
