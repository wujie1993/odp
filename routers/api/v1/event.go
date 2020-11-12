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

type EventController struct {
	controller.BaseController
}

// @summary 获取所有事件
// @tags Event
// @produce json
// @accept json
// @param beginTime query integer false "起始时间"
// @param endTime query integer false "结束时间"
// @param resourceKind query string false "资源类别" Enums(appInstance,host,k8sconfig)
// @param resourceNamespace query string false "资源命名空间"
// @param resourceName query string false "资源标识名称"
// @param action query string false "行为" Enums(Install,Configure,Uninstall,HealthCheck,Label,Connect,Initial)
// @success 200 {object} controller.Response{Data=[]v1.Event}
// @failure 500 {object} controller.Response
// @router /api/v1/events [get]
func (c *EventController) GetEvents(ctx *gin.Context) {
	c.List(ctx, c.listFilt)
}

// @summary 获取单个事件
// @tags Event
// @produce json
// @accept json
// @param name path string true "事件名称"
// @success 200 {object} controller.Response{Data=v1.Event}
// @failure 500 {object} controller.Response
// @router /api/v1/events/{name} [get]
func (c *EventController) GetEvent(ctx *gin.Context) {
	c.Get(ctx)
}

// @summary 创建单个事件
// @tags Event
// @produce json
// @accept json
// @param body body v1.Event true "事件信息"
// @success 200 {object} controller.Response{Data=v1.Event}
// @failure 500 {object} controller.Response
// @router /api/v1/events [post]
func (c *EventController) PostEvent(ctx *gin.Context) {
	c.Create(ctx)
}

// @summary 更新单个事件
// @tags Event
// @produce json
// @accept json
// @param name path string true "事件名称"
// @param body body v1.Event true "事件信息"
// @success 200 {object} controller.Response{Data=v1.Event}
// @failure 500 {object} controller.Response
// @router /api/v1/events/{name} [put]
func (c *EventController) PutEvent(ctx *gin.Context) {
	c.Update(ctx)
}

// @summary 删除单个事件
// @tags Event
// @produce json
// @accept json
// @param name path string true "事件名称"
// @success 200 {object} controller.Response{Data=v1.Event}
// @failure 500 {object} controller.Response
// @router /api/v1/events/{name} [delete]
func (c *EventController) DeleteEvent(ctx *gin.Context) {
	c.Delete(ctx)
}

// 实现了ListFilter的过滤方法
func (c *EventController) listFilt(ctx *gin.Context, objs []core.ApiObject) []core.ApiObject {
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
		event := obj.(*v1.Event)
		// 根据时间过滤
		if event.Metadata.CreateTime.Before(beginTime) || event.Metadata.CreateTime.After(endTime) {
			continue
		}
		// 根据资源过滤
		resourceKind := ctx.Query("resourceKind")
		if resourceKind != "" && event.Spec.ResourceRef.Kind != resourceKind {
			continue
		}
		resourceNamespace := ctx.Query("resourceNamespace")
		if resourceNamespace != "" && event.Spec.ResourceRef.Namespace != resourceNamespace {
			continue
		}
		resourceName := ctx.Query("resourceName")
		if resourceName != "" && event.Spec.ResourceRef.Name != resourceName {
			continue
		}
		// 根据源地址过滤
		action := ctx.Query("action")
		if action != "" && event.Spec.Action != action {
			continue
		}
		result = append(result, obj)
	}

	// 将结果以创建时间降序排序
	sort.Sort(sort.Reverse(core.SortByCreateTime(result)))
	return result
}

func NewEventController() EventController {
	return EventController{
		BaseController: controller.NewController(v1.NewEventRegistry()),
	}
}
