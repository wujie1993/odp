package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/wujie1993/waves/pkg/controller"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

type NamespaceController struct {
	controller.BaseController
}

// @summary 获取所有命名空间
// @tags Namespace
// @produce json
// @accept json
// @success 200 {object} controller.Response{Data=[]v1.Namespace}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces [get]
func (c *NamespaceController) GetNamespaces(ctx *gin.Context) {
	c.List(ctx)
}

// @summary 创建单个命名空间
// @tags Namespace
// @produce json
// @accept json
// @param body body v1.Namespace true "命名空间信息"
// @success 200 {object} controller.Response{Data=v1.Namespace}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces [post]
func (c *NamespaceController) PostNamespace(ctx *gin.Context) {
	c.Create(ctx)
}

// @summary 获取单个命名空间
// @tags Namespace
// @produce json
// @accept json
// @param name path string true "命名空间名称"
// @success 200 {object} controller.Response{Data=v1.Namespace}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces/{name} [get]
func (c *NamespaceController) GetNamespace(ctx *gin.Context) {
	ctx.Set("name", ctx.Param("namespace"))
	c.Get(ctx)
}

// @summary 更新单个命名空间
// @tags Namespace
// @produce json
// @accept json
// @param name path string true "命名空间名称"
// @param body body v1.Namespace true "命名空间信息"
// @success 200 {object} controller.Response{Data=v1.Namespace}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces/{name} [put]
func (c *NamespaceController) PutNamespace(ctx *gin.Context) {
	ctx.Set("name", ctx.Param("namespace"))
	c.Update(ctx)
}

// @summary 删除单个命名空间
// @tags Namespace
// @produce json
// @accept json
// @param name path string true "命名空间名称"
// @success 200 {object} controller.Response{Data=v1.Namespace}
// @failure 500 {object} controller.Response
// @router /api/v1/namespaces/{name} [delete]
func (c *NamespaceController) DeleteNamespace(ctx *gin.Context) {
	ctx.Set("name", ctx.Param("namespace"))
	c.Delete(ctx)
}

func NewNamespaceController() NamespaceController {
	return NamespaceController{
		BaseController: controller.NewController(v1.NewNamespaceRegistry()),
	}
}
