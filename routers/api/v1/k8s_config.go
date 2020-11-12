package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/wujie1993/waves/pkg/controller"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

type K8sConfigController struct {
	controller.BaseController
}

// @summary 获取所有k8s集群
// @tags K8sConfig
// @produce json
// @accept json
// @param namespace path string true "命名空间" default(default)
// @success 200 {object} controller.Response{Data=[]v1.K8sConfig}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces/{namespace}/k8sconfig [get]
func (c *K8sConfigController) GetK8ClusterConfigs(ctx *gin.Context) {
	c.List(ctx)
}

// @Summary 获取单个k8s集群配置
// @tags K8sConfig
// @Produce json
// @Accept  json
// @param namespace path string true "命名空间" default(default)
// @param name path string true "集群名称"
// @success 200 {object} controller.Response{Data=v1.K8sConfig}
// @Failure 500 {object} controller.Response
// @Router /api/v1/namespaces/{namespace}/k8sconfig/{name} [get]
func (c *K8sConfigController) GetK8ClusterConfig(ctx *gin.Context) {
	c.Get(ctx)
}

// @Summary 创建单个k8s集群配置
// @tags K8sConfig
// @Produce json
// @Accept  json
// @param namespace path string true "命名空间" default(default)
// @Param body body v1.K8sConfig true "k8s集群信息"
// @Success 200 {object} controller.Response{Data=v1.K8sConfig}
// @Failure 500 {object} controller.Response
// @Router /api/v1/namespaces/{namespace}/k8sconfig [post]
func (c *K8sConfigController) PostK8ClusterConfig(ctx *gin.Context) {
	c.Create(ctx)
}

// @Summary 更新单个k8s集群配置
// @tags K8sConfig
// @Produce json
// @Accept  json
// @param namespace path string true "命名空间" default(default)
// @param name path string true "集群名称"
// @Param body body v1.K8sConfig true "k8s集群信息"
// @Success 200 {object} controller.Response{Data=v1.K8sConfig}
// @Failure 500 {object} controller.Response
// @Router /api/v1/namespaces/{namespace}/k8sconfig/{name} [put]
func (c *K8sConfigController) PutK8ClusterConfig(ctx *gin.Context) {
	c.Update(ctx)
}

// @summary 删除单个k8s集群配置
// @tags K8sConfig
// @produce json
// @accept json
// @param namespace path string true "命名空间"
// @param name path string true "集群名称"
// @success 200 {object} controller.Response{Data=v1.K8sConfig}
// @failure 500 {object} controller.Response
// @Router /api/v1/namespaces/{namespace}/k8sconfig/{name} [delete]
func (c *K8sConfigController) DeleteK8sClusterConfig(ctx *gin.Context) {
	c.Delete(ctx)
}

func NewK8sConfigController() K8sConfigController {
	return K8sConfigController{
		BaseController: controller.NewController(v1.NewK8sConfigRegistry()),
	}
}
