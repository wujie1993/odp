package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/wujie1993/waves/pkg/controller"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

type HostController struct {
	controller.BaseController
}

// @summary 获取所有主机
// @tags Host
// @produce json
// @accept json
// @success 200 {object} controller.Response{Data=[]v1.Host}
// @failure 500 {object} controller.Response
// @router /api/v1/hosts [get]
func (c *HostController) GetHosts(ctx *gin.Context) {
	c.List(ctx)
}

// @summary 获取单个主机
// @tags Host
// @produce json
// @accept json
// @param name path string true "主机名称"
// @success 200 {object} controller.Response{Data=v1.Host}
// @failure 500 {object} controller.Response
// @router /api/v1/hosts/{name} [get]
func (c *HostController) GetHost(ctx *gin.Context) {
	c.Get(ctx)
}

// @summary 创建单个主机
// @tags Host
// @produce json
// @accept json
// @param body body v1.Host true "主机信息"
// @success 200 {object} controller.Response{Data=v1.Host}
// @failure 500 {object} controller.Response
// @router /api/v1/hosts [post]
func (c *HostController) PostHost(ctx *gin.Context) {
	c.Create(ctx)
}

// @summary 更新单个主机
// @tags Host
// @produce json
// @accept json
// @param name path string true "主机名称"
// @param body body v1.Host true "主机信息"
// @success 200 {object} controller.Response{Data=v1.Host}
// @failure 500 {object} controller.Response
// @router /api/v1/hosts/{name} [put]
func (c *HostController) PutHost(ctx *gin.Context) {
	c.Update(ctx)
}

// @summary 删除单个主机
// @tags Host
// @produce json
// @accept json
// @param name path string true "主机名称"
// @success 200 {object} controller.Response{Data=v1.Host}
// @failure 500 {object} controller.Response
// @router /api/v1/hosts/{name} [delete]
func (c *HostController) DeleteHost(ctx *gin.Context) {
	c.Delete(ctx)
}

func NewHostController() HostController {
	return HostController{
		BaseController: controller.NewController(v1.NewHostRegistry()),
	}
}
