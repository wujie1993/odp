package wavectl

import (
	"context"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"

	clientset "github.com/wujie1993/waves/pkg/client"
	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

type AppClient struct {
	clientset.ClientSet
}

func (c AppClient) GetPrint(namespace string, name string, format string) error {
	if name != "" {
		app, err := c.ClientSet.V1().Apps("default").Get(context.TODO(), name)
		if err != nil {
			log.Error(err)
			return err
		}
		return printApps([]v1.App{*app}, format)
	}

	apps, err := c.ClientSet.V1().Apps("default").List(context.TODO())
	if err != nil {
		log.Error(err)
		return err
	}
	return printApps(apps, format)
}

func (c AppClient) Apply(obj core.ApiObject) (core.ApiObject, error) {
	return nil, e.Errorf("unsupported function")
}

func (c AppClient) Create(obj core.ApiObject) (core.ApiObject, error) {
	return nil, e.Errorf("unsupported function")
}

func (c AppClient) Update(obj core.ApiObject) (core.ApiObject, error) {
	return nil, e.Errorf("unsupported function")
}

func (c AppClient) Delete(namespace string, name string) (core.ApiObject, error) {
	return nil, e.Errorf("unsupported function")
}

func printApps(apps []v1.App, format string) error {
	switch format {
	case OutputFormatJSON:
		data, err := ToJSON(apps, false)
		if err != nil {
			log.Error(err)
			return err
		}
		fmt.Println(string(data))
	case OutputFormatJSONPretty:
		data, err := ToJSON(apps, true)
		if err != nil {
			log.Error(err)
			return err
		}
		fmt.Println(string(data))
	case OutputFormatYAML:
		for _, app := range apps {
			data, err := app.ToYAML()
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
		table.SetHeader([]string{"name", "alias", "category", "latest version", "versions"})

		for _, app := range apps {
			totalVersion := len(app.Spec.Versions)
			table.Append([]string{
				app.Metadata.Name,
				app.Spec.Versions[totalVersion-1].Desc,
				core.GetCategoryMsg(app.Spec.Category),
				app.Spec.Versions[totalVersion-1].Version,
				fmt.Sprint(totalVersion),
			})
		}
		table.Render()
	}
	return nil
}
