package v1

import (
	"sort"

	"github.com/gin-gonic/gin"

	"github.com/wujie1993/waves/pkg/controller"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

type RevisionController struct {
	controller.BaseController
}

// @summary 获取所有修订历史
// @tags Revision
// @produce json
// @accept json
// @success 200 {object} controller.Response{Data=[]v1.Revision}
// @failure 500 {object} controller.Response
// @router /api/v1/revisions [get]
func (c *RevisionController) GetRevisions(ctx *gin.Context) {
	c.List(ctx, c.listFilt)
}

// @summary 获取单个修订历史
// @tags Revision
// @produce json
// @accept json
// @param name path string true "修订历史名称"
// @success 200 {object} controller.Response{Data=v1.Revision}
// @failure 500 {object} controller.Response
// @router /api/v1/revisions/{name} [get]
func (c *RevisionController) GetRevision(ctx *gin.Context) {
	c.Get(ctx)
}

// @summary 创建单个修订历史
// @tags Revision
// @produce json
// @accept json
// @param body body v1.Revision true "修订历史信息"
// @success 200 {object} controller.Response{Data=v1.Revision}
// @failure 500 {object} controller.Response
// @router /api/v1/revisions [post]
func (c *RevisionController) PostRevision(ctx *gin.Context) {
	c.Create(ctx)
}

// @summary 更新单个修订历史
// @tags Revision
// @produce json
// @accept json
// @param name path string true "修订历史名称"
// @param body body v1.Revision true "修订历史信息"
// @success 200 {object} controller.Response{Data=v1.Revision}
// @failure 500 {object} controller.Response
// @router /api/v1/revisions/{name} [put]
func (c *RevisionController) PutRevision(ctx *gin.Context) {
	c.Update(ctx)
}

// @summary 删除单个修订历史
// @tags Revision
// @produce json
// @accept json
// @param name path string true "修订历史名称"
// @success 200 {object} controller.Response{Data=v1.Revision}
// @failure 500 {object} controller.Response
// @router /api/v1/revisions/{name} [delete]
func (c *RevisionController) DeleteRevision(ctx *gin.Context) {
	c.Delete(ctx)
}

// 实现了ListFilter的过滤方法
func (c *RevisionController) listFilt(ctx *gin.Context, objs []core.ApiObject) []core.ApiObject {
	sort.Sort(core.SortByCreateTime(objs))
	return objs
}

func NewRevisionController() RevisionController {
	return RevisionController{
		BaseController: controller.NewController(v1.NewRevisionRegistry()),
	}
}
