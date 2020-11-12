package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/wujie1993/waves/pkg/controller"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

type PkgController struct {
	controller.BaseController
}

// @summary 获取所有部署包
// @tags Pkg
// @produce json
// @accept json
// @success 200 {object} controller.Response{Data=[]v1.Pkg}
// @failure 500 {object} controller.Response
// @router /api/v1/pkgs [get]
func (c *PkgController) GetPkgs(ctx *gin.Context) {
	c.List(ctx)
}

// @summary 获取单个部署包
// @tags Pkg
// @produce json
// @accept json
// @param name path string true "部署包名称"
// @success 200 {object} controller.Response{Data=v1.Pkg}
// @failure 500 {object} controller.Response
// @router /api/v1/pkgs/{name} [get]
func (c *PkgController) GetPkg(ctx *gin.Context) {
	c.Get(ctx)
}

// @summary 创建单个部署包
// @tags Pkg
// @produce json
// @accept json
// @param body body v1.Pkg true "部署包信息"
// @success 200 {object} controller.Response{Data=v1.Pkg}
// @failure 500 {object} controller.Response
// @router /api/v1/pkgs [post]
func (c *PkgController) PostPkg(ctx *gin.Context) {
	c.Create(ctx)
}

// @summary 更新单个部署包
// @tags Pkg
// @produce json
// @accept json
// @param name path string true "部署包名称"
// @param body body v1.Pkg true "部署包信息"
// @success 200 {object} controller.Response{Data=v1.Pkg}
// @failure 500 {object} controller.Response
// @router /api/v1/pkgs/{name} [put]
func (c *PkgController) PutPkg(ctx *gin.Context) {
	c.Update(ctx)
}

// @summary 删除单个部署包
// @tags Pkg
// @produce json
// @accept json
// @param name path string true "部署包名称"
// @success 200 {object} controller.Response{Data=v1.Pkg}
// @failure 500 {object} controller.Response
// @router /api/v1/pkgs/{name} [delete]
func (c *PkgController) DeletePkg(ctx *gin.Context) {
	c.Delete(ctx)
}

func NewPkgController() PkgController {
	return PkgController{
		BaseController: controller.NewController(v1.NewPkgRegistry()),
	}
}
