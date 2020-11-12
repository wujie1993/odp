## 路由注册

### 为实体对象注册路由

以注册主机路由为例

1. 在`../main.go`中为主机添加接口文档标签注释

   ```
   ...
   // @tag.name Host
   // @tag.description 主机
   ...
   ```

2. 在`./api/v1/host.go`添加主机控制器

   ```
   package v1
   
   import (
   	"github.com/gin-gonic/gin"
   
   	"github.com/wujie1993/waves/pkg/orm"
   	"github.com/wujie1993/waves/pkg/orm/v1"
   )
   
   type V1HostController struct {
   	V1BaseController
   }
   
   // @summary 获取所有主机
   // @tags Host
   // @produce json
   // @accept json
   // @success 200 {object} Response{Data=[]v1.V1Host}
   // @failure 500 {object} Response
   // @router /api/v1/hosts [get]
   func (c *V1HostController) GetHosts(ctx *gin.Context) {
   	c.list(ctx)
   }
   
   // @summary 获取单个主机
   // @tags Host
   // @produce json
   // @accept json
   // @param name path string true "主机名称"
   // @success 200 {object} Response{Data=v1.V1Host}
   // @failure 500 {object} Response
   // @router /api/v1/hosts/{name} [get]
   func (c *V1HostController) GetHost(ctx *gin.Context) {
   	c.get(ctx)
   }
   
   // @summary 创建单个主机
   // @tags Host
   // @produce json
   // @accept json
   // @param body body v1.V1Host true "主机信息"
   // @success 200 {object} Response{Data=v1.V1Host}
   // @failure 500 {object} Response
   // @router /api/v1/hosts [post]
   func (c *V1HostController) PostHost(ctx *gin.Context) {
   	c.create(ctx)
   }
   
   // @summary 更新单个主机
   // @tags Host
   // @produce json
   // @accept json
   // @param name path string true "主机名称"
   // @param body body v1.V1Host true "主机信息"
   // @success 200 {object} Response{Data=v1.V1Host}
   // @failure 500 {object} Response
   // @router /api/v1/hosts/{name} [put]
   func (c *V1HostController) PutHost(ctx *gin.Context) {
   	c.update(ctx)
   }
   
   // @summary 删除单个主机
   // @tags Host
   // @produce json
   // @accept json
   // @param name path string true "主机名称"
   // @success 200 {object} Response{Data=v1.V1Host}
   // @failure 500 {object} Response
   // @router /api/v1/hosts/{name} [delete]
   func (c *V1HostController) DeleteHost(ctx *gin.Context) {
   	c.delete(ctx)
   }
   
   func NewV1HostController() V1HostController {
   	return V1HostController{
   		V1BaseController: V1BaseController{
   			helper:     orm.GetHelper(),
   			V1Registry: v1.NewV1HostRegistry().V1Registry,
   		},
   	}
   }
   ```

3. 在`./router.go`中为主机控制器添加路由注册

```
apiv1 := r.Group(setting.AppSetting.PrefixUrl + "/api/v1")
{
        ...
        
	host := apiv1.Group("/hosts")
	{
		c := v1.NewV1HostController()
		host.GET("", c.GetHosts)
		host.POST("", c.PostHost)
		host.GET(":name", c.GetHost)
	        host.PUT(":name", c.PutHost)
		host.DELETE(":name", c.DeleteHost)
	}
        ...
}
```

### 查询过滤

以审计日志为例

通常情况下通过`GET /api/v1/audits`接口可获取到所有的审计日志，但当审计日志数量众多时，返回所有的结果会使查询效率降低，且不便于客户端进行数据过滤，因此需要查询接口可以接收各种过滤参数

数据过滤可通过向`V1BaseController.list(ctx *gin.Context, filts ...ListFilter)`中传递多个实现了`ListFilter`的过滤方法，过滤方法的实现如下：

```
// 实现过滤方法
// type ListFilter func(*gin.Context, []core.ApiObject) []core.ApiObject
func (c *V1AuditController) listFilt(ctx *gin.Context, objs []core.ApiObject) []core.ApiObject {
        // TODO: 将objs过滤并返回
}
```

具体实现可以参照[审计日志](./api/v1/audit.go)
