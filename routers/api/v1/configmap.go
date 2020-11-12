package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/wujie1993/waves/pkg/controller"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

type ConfigMapController struct {
	controller.BaseController
}

// @summary 获取所有配置字典
// @tags ConfigMap
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @success 200 {object} controller.Response{Data=[]v1.ConfigMap}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces/{namespace}/configmaps [get]
func (c *ConfigMapController) GetConfigMaps(ctx *gin.Context) {
	c.List(ctx)
}

// @summary 获取单个配置字典
// @tags ConfigMap
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param name path string true "配置字典名称"
// @success 200 {object} controller.Response{Data=v1.ConfigMap}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces/{namespace}/configmaps/{name} [get]
func (c *ConfigMapController) GetConfigMap(ctx *gin.Context) {
	c.Get(ctx)
}

// @summary 创建单个配置字典
// @tags ConfigMap
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param body body v1.ConfigMap true "配置字典信息"
// @success 200 {object} controller.Response{Data=v1.ConfigMap}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces/{namespace}/configmaps [post]
func (c *ConfigMapController) PostConfigMap(ctx *gin.Context) {
	c.Create(ctx)
}

// @summary 更新单个配置字典
// @tags ConfigMap
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param name path string true "配置字典名称"
// @param body body v1.ConfigMap true "配置字典信息"
// @success 200 {object} controller.Response{Data=v1.ConfigMap}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces/{namespace}/configmaps/{name} [put]
func (c *ConfigMapController) PutConfigMap(ctx *gin.Context) {
	c.Update(ctx)
}

// @summary 删除单个配置字典
// @tags ConfigMap
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @param name path string true "配置字典名称"
// @success 200 {object} controller.Response{Data=v1.ConfigMap}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces/{namespace}/configmaps/{name} [delete]
func (c *ConfigMapController) DeleteConfigMap(ctx *gin.Context) {
	c.Delete(ctx)
}

func NewConfigMapController() ConfigMapController {
	return ConfigMapController{
		BaseController: controller.NewController(v1.NewConfigMapRegistry()),
	}
}
