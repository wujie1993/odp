package v2

import (
	"github.com/gin-gonic/gin"

	"github.com/wujie1993/waves/pkg/controller"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v2"
)

type AppInstanceController struct {
	controller.BaseController
}

// @summary 获取所有应用实例
// @tags AppInstance
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param category query string false "应用分类" Enums(thirdParty,customize,hostPlugin,algorithmPlugin)
// @success 200 {object} controller.Response{Data=[]v2.AppInstance}
// @failure 500 {object} controller.Response
// @router /api/v2/namespaces/{namespace}/appinstances [get]
func (c *AppInstanceController) GetAppInstances(ctx *gin.Context) {
	c.List(ctx, c.listFilt)
}

// @summary 获取单个应用实例
// @tags AppInstance
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param name path string true "应用实例名称"
// @success 200 {object} controller.Response{Data=v2.AppInstance}
// @failure 500 {object} controller.Response
// @router /api/v2/namespaces/{namespace}/appinstances/{name} [get]
func (c *AppInstanceController) GetAppInstance(ctx *gin.Context) {
	c.Get(ctx)
}

// @summary 创建单个应用实例
// @tags AppInstance
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param body body v2.AppInstance true "应用实例信息"
// @success 200 {object} controller.Response{Data=v2.AppInstance}
// @failure 500 {object} controller.Response
// @router /api/v2/namespaces/{namespace}/appinstances [post]
func (c *AppInstanceController) PostAppInstance(ctx *gin.Context) {
	c.Create(ctx)
}

// @summary 更新单个应用实例
// @tags AppInstance
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param name path string true "应用实例名称"
// @param body body v2.AppInstance true "应用实例信息"
// @success 200 {object} controller.Response{Data=v2.AppInstance}
// @failure 500 {object} controller.Response
// @router /api/v2/namespaces/{namespace}/appinstances/{name} [put]
func (c *AppInstanceController) PutAppInstance(ctx *gin.Context) {
	c.Update(ctx)
}

// @summary 删除单个应用实例
// @tags AppInstance
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param name path string true "应用实例名称"
// @success 200 {object} controller.Response{Data=v2.AppInstance}
// @failure 500 {object} controller.Response
// @router /api/v2/namespaces/{namespace}/appinstances/{name} [delete]
func (c *AppInstanceController) DeleteAppInstance(ctx *gin.Context) {
	c.Delete(ctx)
}

// @summary 获取单个应用实例的所有修订版本
// @tags AppInstance
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param name path string true "应用实例名称"
// @success 200 {object} controller.Response{Data=[]v2.AppInstance}
// @failure 500 {object} controller.Response
// @router /api/v2/namespaces/{namespace}/appinstances/{name}/revisions [get]
func (c *AppInstanceController) GetAppInstanceRevisions(ctx *gin.Context) {
	c.ListRevisions(ctx)
}

// @summary 获取单个应用实例的指定修订版本
// @tags AppInstance
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param name path string true "应用实例名称"
// @param revision path integer true "修订版本"
// @success 200 {object} controller.Response{Data=v2.AppInstance}
// @failure 500 {object} controller.Response
// @router /api/v2/namespaces/{namespace}/appinstances/{name}/revisions/{revision} [get]
func (c *AppInstanceController) GetAppInstanceRevision(ctx *gin.Context) {
	c.GetRevision(ctx)
}

// @summary 更新单个应用实例到指定的修订版本
// @tags AppInstance
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param name path string true "应用实例名称"
// @param revision path integer true "修订版本"
// @success 200 {object} controller.Response{Data=v2.AppInstance}
// @failure 500 {object} controller.Response
// @router /api/v2/namespaces/{namespace}/appinstances/{name}/revisions/{revision} [put]
func (c *AppInstanceController) PutAppInstanceRevision(ctx *gin.Context) {
	c.PutRevision(ctx)
}

// @summary 删除单个应用实例的指定修订版本
// @tags AppInstance
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param name path string true "应用实例名称"
// @param revision path integer true "修订版本"
// @success 200 {object} controller.Response{Data=v2.AppInstance}
// @failure 500 {object} controller.Response
// @router /api/v2/namespaces/{namespace}/appinstances/{name}/revisions/{revision} [delete]
func (c *AppInstanceController) DeleteAppInstanceRevision(ctx *gin.Context) {
	c.DeleteRevision(ctx)
}

// 实现了ListFilter的过滤方法
func (c *AppInstanceController) listFilt(ctx *gin.Context, objs []core.ApiObject) []core.ApiObject {
	result := []core.ApiObject{}

	category := ctx.Query("category")

	for _, obj := range objs {
		appInstance := obj.(*v2.AppInstance)
		if category != "" && appInstance.Spec.Category != category {
			continue
		}
		result = append(result, obj)
	}

	return result
}

func NewAppInstanceController() AppInstanceController {
	c := AppInstanceController{
		BaseController: controller.NewController(v2.NewAppInstanceRegistry()),
	}
	c.SetRevisioner(v2.NewAppInstanceRevision())
	return c
}
