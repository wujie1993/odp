package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/wujie1993/waves/pkg/controller"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

type GPUController struct {
	controller.BaseController
}

// @summary 获取所有显卡
// @tags GPU
// @produce json
// @accept json
// @param category query string false "显卡分类" Enums(thirdParty,customize,hostPlugin,algorithmPlugin,algorithmInstance)
// @success 200 {object} controller.Response{Data=[]v1.GPU}
// @failure 500 {object} controller.Response
// @router /api/v1/gpus [get]
func (c *GPUController) GetGPUs(ctx *gin.Context) {
	c.List(ctx)
}

// @summary 获取单个显卡
// @tags GPU
// @produce json
// @accept json
// @param name path string true "显卡名称"
// @success 200 {object} controller.Response{Data=v1.GPU}
// @failure 500 {object} controller.Response
// @router /api/v1/gpus/{name} [get]
func (c *GPUController) GetGPU(ctx *gin.Context) {
	c.Get(ctx)
}

// @summary 创建单个显卡
// @tags GPU
// @produce json
// @accept json
// @param body body v1.GPU true "显卡信息"
// @success 200 {object} controller.Response{Data=v1.GPU}
// @failure 500 {object} controller.Response
// @router /api/v1/gpus [post]
func (c *GPUController) PostGPU(ctx *gin.Context) {
	c.Create(ctx)
}

// @summary 更新单个显卡
// @tags GPU
// @produce json
// @accept json
// @param name path string true "显卡名称"
// @param body body v1.GPU true "显卡信息"
// @success 200 {object} controller.Response{Data=v1.GPU}
// @failure 500 {object} controller.Response
// @router /api/v1/gpus/{name} [put]
func (c *GPUController) PutGPU(ctx *gin.Context) {
	c.Update(ctx)
}

// @summary 删除单个显卡
// @tags GPU
// @produce json
// @accept json
// @param name path string true "显卡名称"
// @success 200 {object} controller.Response{Data=v1.GPU}
// @failure 500 {object} controller.Response
// @router /api/v1/gpus/{name} [delete]
func (c *GPUController) DeleteGPU(ctx *gin.Context) {
	c.Delete(ctx)
}

func NewGPUController() GPUController {
	return GPUController{
		BaseController: controller.NewController(v1.NewGPURegistry()),
	}
}
