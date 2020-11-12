package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

type ListFilter func(*gin.Context, []core.ApiObject) []core.ApiObject

type Response struct {
	OpCode int         `json:"OpCode"`
	OpDesc string      `json:"OpDesc"`
	Data   interface{} `json:"Data"`
}

type BaseController struct {
	registry registry.ApiObjectRegistry
	helper   orm.Helper
}

func (c *BaseController) Response(ctx *gin.Context, httpCode, opCode int, opMsg string, data interface{}) {
	resp := Response{
		OpCode: opCode,
		OpDesc: opMsg,
		Data:   data,
	}

	c.recordAudit(ctx, httpCode, resp)

	ctx.JSON(httpCode, resp)
}

func (c *BaseController) List(ctx *gin.Context, filts ...ListFilter) {
	namespace := ctx.Param("namespace")

	result, err := c.registry.List(context.TODO(), namespace)
	if err != nil {
		log.Error(err)
		c.Response(ctx, 500, e.ERROR, err.Error(), nil)
		return
	}

	for _, filt := range filts {
		result = filt(ctx, result)
	}

	c.Response(ctx, 200, e.SUCCESS, "", result)
}

func (c *BaseController) Get(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	result, err := c.registry.Get(context.TODO(), namespace, name)
	if err != nil {
		log.Error(err)
		c.Response(ctx, 500, e.ERROR, err.Error(), nil)
		return
	}

	c.Response(ctx, 200, e.SUCCESS, "", result)
}

func (c *BaseController) Create(ctx *gin.Context) {
	obj, err := orm.New(c.registry.GVK())
	if err != nil {
		log.Error(err)
		c.Response(ctx, 400, e.ERROR, err.Error(), nil)
		return
	}

	if err := ctx.ShouldBindBodyWith(obj, binding.JSON); err != nil {
		log.Error(err)
		c.Response(ctx, 400, e.INVALID_PARAMS, err.Error(), nil)
		return
	}

	result, err := c.registry.Create(context.TODO(), obj)
	if err != nil {
		log.Error(err)
		c.Response(ctx, 500, e.ERROR, err.Error(), nil)
		return
	}

	c.Response(ctx, 200, e.SUCCESS, "", result)
}

func (c *BaseController) Update(ctx *gin.Context) {
	obj, err := orm.New(c.registry.GVK())
	if err != nil {
		log.Error(err)
		c.Response(ctx, 500, e.ERROR, err.Error(), nil)
		return
	}

	if err := ctx.ShouldBindBodyWith(obj, binding.JSON); err != nil {
		log.Error(err)
		c.Response(ctx, 400, e.INVALID_PARAMS, err.Error(), nil)
		return
	}

	result, err := c.registry.Update(context.TODO(), obj)
	if err != nil {
		log.Error(err)
		c.Response(ctx, 500, e.ERROR, err.Error(), nil)
		return
	}

	c.Response(ctx, 200, e.SUCCESS, "", result)
}

func (c *BaseController) Delete(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	result, err := c.registry.Delete(context.TODO(), namespace, name)
	if err != nil {
		log.Error(err)
		c.Response(ctx, 500, e.ERROR, err.Error(), nil)
		return
	}

	c.Response(ctx, 200, e.SUCCESS, "", result)
}

func (c *BaseController) recordAudit(ctx *gin.Context, httpCode int, resp Response) {
	audit := v1.NewAudit()

	switch ctx.Request.Method {
	case http.MethodPost:
		audit.Spec.Action = core.AuditActionCreate
	case http.MethodPut:
		audit.Spec.Action = core.AuditActionUpdate
	case http.MethodDelete:
		audit.Spec.Action = core.AuditActionDelete
	default:
		return
	}

	if ctx.Request.Method == http.MethodPut || ctx.Request.Method == http.MethodPost {
		reqObj, err := orm.New(c.registry.GVK())
		if err != nil {
			log.Error(err)
			return
		}
		if err := ctx.ShouldBindBodyWith(reqObj, binding.JSON); err != nil {
			log.Error(err)
			return
		}
		reqBodyData, err := json.Marshal(reqObj)
		if err != nil {
			log.Error(err)
			return
		}
		audit.Spec.ReqBody = string(reqBodyData)
	}
	if resp.Data != nil {
		if data, err := json.Marshal(resp.Data); err != nil {
			log.Error(err)
			return
		} else {
			audit.Spec.RespBody = string(data)
		}
	}
	respObj, ok := resp.Data.(core.ApiObject)
	if ok {
		metadata := respObj.GetMetadata()
		audit.Metadata.Annotations["ShortName"] = metadata.Annotations["ShortName"]
		audit.Spec.ResourceRef = v1.ResourceRef{
			Kind:      c.registry.GVK().Kind,
			Name:      metadata.Name,
			Namespace: metadata.Namespace,
		}
	} else {
		audit.Spec.ResourceRef = v1.ResourceRef{
			Kind:      c.registry.GVK().Kind,
			Name:      ctx.Param("name"),
			Namespace: ctx.Param("namespace"),
		}
	}
	audit.Spec.SourceIP = ctx.ClientIP()
	audit.Spec.StatusCode = httpCode
	audit.Spec.Msg = resp.OpDesc

	if err := c.helper.V1.Audit.Record(audit); err != nil {
		log.Error(err)
	}
}

func NewController(registry registry.ApiObjectRegistry) BaseController {
	return BaseController{
		helper:   orm.GetHelper(),
		registry: registry,
	}
}
