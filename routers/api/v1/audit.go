package v1

import (
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/wujie1993/waves/pkg/controller"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

type AuditController struct {
	controller.BaseController
}

// @summary 获取所有审计
// @tags Audit
// @produce json
// @accept json
// @param beginTime query integer false "起始时间"
// @param endTime query integer false "结束时间"
// @param resourceKind query string false "资源类别" Enums(app,appInstance,audit,event,host,job,configMap,k8sconfig)
// @param resourceNamespace query string false "资源命名空间"
// @param resourceName query string false "资源标识名称"
// @param sourceIP query string false "来源地址"
// @success 200 {object} controller.Response{Data=[]v1.Audit}
// @failure 500 {object} controller.Response
// @router /api/v1/audits [get]
func (c *AuditController) GetAudits(ctx *gin.Context) {
	c.List(ctx, c.listFilt)
}

// @summary 获取单个审计
// @tags Audit
// @produce json
// @accept json
// @param name path string true "审计名称"
// @success 200 {object} controller.Response{Data=v1.Audit}
// @failure 500 {object} controller.Response
// @router /api/v1/audits/{name} [get]
func (c *AuditController) GetAudit(ctx *gin.Context) {
	c.Get(ctx)
}

// @summary 创建单个审计
// @tags Audit
// @produce json
// @accept json
// @param body body v1.Audit true "审计信息"
// @success 200 {object} controller.Response{Data=v1.Audit}
// @failure 500 {object} controller.Response
// @router /api/v1/audits [post]
func (c *AuditController) PostAudit(ctx *gin.Context) {
	c.Create(ctx)
}

// @summary 更新单个审计
// @tags Audit
// @produce json
// @accept json
// @param name path string true "审计名称"
// @param body body v1.Audit true "审计信息"
// @success 200 {object} controller.Response{Data=v1.Audit}
// @failure 500 {object} controller.Response
// @router /api/v1/audits/{name} [put]
func (c *AuditController) PutAudit(ctx *gin.Context) {
	c.Update(ctx)
}

// @summary 删除单个审计
// @tags Audit
// @produce json
// @accept json
// @param name path string true "审计名称"
// @success 200 {object} controller.Response{Data=v1.Audit}
// @failure 500 {object} controller.Response
// @router /api/v1/audits/{name} [delete]
func (c *AuditController) DeleteAudit(ctx *gin.Context) {
	c.Delete(ctx)
}

// 实现了ListFilter的过滤方法
func (c *AuditController) listFilt(ctx *gin.Context, objs []core.ApiObject) []core.ApiObject {
	result := []core.ApiObject{}
	var beginTime, endTime time.Time

	beginTimeStr := ctx.Query("beginTime")
	// 默认起始时间为一周前
	if beginTimeStr == "" {
		beginTime = time.Now().AddDate(0, 0, -7)
	} else {
		beginTimestamp, err := strconv.ParseInt(beginTimeStr, 10, 64)
		if err != nil {
			beginTime = time.Now().AddDate(0, 0, -7)
		}
		beginTime = time.Unix(beginTimestamp, 0)
	}
	endTimeStr := ctx.Query("endTime")
	// 默认结束时间为当前时间
	if endTimeStr == "" {
		endTime = time.Now()
	} else {
		endTimestamp, err := strconv.ParseInt(endTimeStr, 10, 64)
		if err != nil {
			endTime = time.Now()
		}
		endTime = time.Unix(endTimestamp, 0)
	}
	if beginTime.After(endTime) {
		return result
	}

	for _, obj := range objs {
		audit := obj.(*v1.Audit)
		// 根据时间过滤
		if audit.Metadata.CreateTime.Before(beginTime) || audit.Metadata.CreateTime.After(endTime) {
			continue
		}
		// 根据资源过滤
		resourceKind := ctx.Query("resourceKind")
		if resourceKind != "" && audit.Spec.ResourceRef.Kind != resourceKind {
			continue
		}
		resourceNamespace := ctx.Query("resourceNamespace")
		if resourceNamespace != "" && audit.Spec.ResourceRef.Namespace != resourceNamespace {
			continue
		}
		resourceName := ctx.Query("resourceName")
		if resourceName != "" && audit.Spec.ResourceRef.Name != resourceName {
			continue
		}
		// 根据源地址过滤
		sourceIP := ctx.Query("sourceIP")
		if sourceIP != "" && audit.Spec.SourceIP != sourceIP {
			continue
		}
		result = append(result, obj)
	}

	// 将结果以创建时间降序排序
	sort.Sort(sort.Reverse(core.SortByCreateTime(result)))
	return result
}

func NewAuditController() AuditController {
	return AuditController{
		BaseController: controller.NewController(v1.NewAuditRegistry()),
	}
}
