package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/wujie1993/waves/pkg/controller"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

type AppController struct {
	controller.BaseController
}

// @summary 获取所有应用
// @tags App
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param category query string false "应用分类" Enums(thirdParty,customize,hostPlugin,algorithmPlugin)
// @success 200 {object} controller.Response{Data=[]v1.App}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces/{namespace}/apps [get]
func (c *AppController) GetApps(ctx *gin.Context) {
	c.List(ctx, c.listFilt)
}

// @summary 获取单个应用
// @tags App
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param name path string true "应用名称"
// @success 200 {object} controller.Response{Data=v1.App}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces/{namespace}/apps/{name} [get]
func (c *AppController) GetApp(ctx *gin.Context) {
	c.Get(ctx)
}

// @summary 创建单个应用
// @tags App
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param body body v1.App true "应用信息"
// @success 200 {object} controller.Response{Data=v1.App}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces/{namespace}/apps [post]
func (c *AppController) PostApp(ctx *gin.Context) {
	c.Create(ctx)
}

// @summary 更新单个应用
// @tags App
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param name path string true "应用名称"
// @param body body v1.App true "应用信息"
// @success 200 {object} controller.Response{Data=v1.App}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces/{namespace}/apps/{name} [put]
func (c *AppController) PutApp(ctx *gin.Context) {
	c.Update(ctx)
}

// @summary 删除单个应用
// @tags App
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param name path string true "应用名称"
// @success 200 {object} controller.Response{Data=v1.App}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces/{namespace}/apps/{name} [delete]
func (c *AppController) DeleteApp(ctx *gin.Context) {
	c.Delete(ctx)
}

// 实现了ListFilter的过滤方法
func (c *AppController) listFilt(ctx *gin.Context, objs []core.ApiObject) []core.ApiObject {
	result := []core.ApiObject{}

	category := ctx.Query("category")

	for _, obj := range objs {
		app := obj.(*v1.App)
		if category != "" && app.Spec.Category != category {
			continue
		}
		result = append(result, obj)
	}

	return result
}

func NewAppController() AppController {
	return AppController{
		BaseController: controller.NewController(v1.NewAppRegistry()),
	}
}
