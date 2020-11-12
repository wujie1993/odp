package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/wujie1993/waves/pkg/controller"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

type ProjectController struct {
	controller.BaseController
}

// @summary 获取所有项目空间
// @tags Project
// @produce json
// @accept json
// @success 200 {object} controller.Response{Data=[]v1.Project}
// @failure 500 {object} controller.Response
// @router /api/v1/project [get]
func (c *ProjectController) GetProjects(ctx *gin.Context) {
	c.List(ctx)
}

// @summary 创建单个项目空间
// @tags Project
// @produce json
// @accept json
// @param body body v1.Project true "项目空间信息"
// @success 200 {object} controller.Response{Data=v1.Project}
// @failure 500 {object} controller.Response
// @router /api/v1/project [post]
func (c *ProjectController) PostProject(ctx *gin.Context) {
	c.Create(ctx)
}

// @summary 获取单个项目空间
// @tags Project
// @produce json
// @accept json
// @param name path string true "项目空间名称"
// @success 200 {object} controller.Response{Data=v1.Project}
// @failure 500 {object} controller.Response
// @router /api/v1/project/{name} [get]
func (c *ProjectController) GetProject(ctx *gin.Context) {
	c.Get(ctx)
}

// @summary 更新单个项目空间
// @tags Project
// @produce json
// @accept json
// @param name path string true "项目空间名称"
// @param body body v1.Project true "项目空间信息"
// @success 200 {object} controller.Response{Data=v1.Project}
// @failure 500 {object} controller.Response
// @router /api/v1/project/{name} [put]
func (c *ProjectController) PutProject(ctx *gin.Context) {
	c.Update(ctx)
}

// @summary 删除单个项目空间
// @tags Project
// @produce json
// @accept json
// @param name path string true "项目空间名称"
// @success 200 {object} controller.Response{Data=v1.Project}
// @failure 500 {object} controller.Response
// @router /api/v1/project/{name} [delete]
func (c *ProjectController) DeleteProject(ctx *gin.Context) {
	c.Delete(ctx)
}

func NewProjectController() ProjectController {
	return ProjectController{
		BaseController: controller.NewController(v1.NewProjectRegistry()),
	}
}
