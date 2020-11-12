package routers

import (
	"github.com/gin-gonic/gin"

	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	_ "github.com/wujie1993/waves/docs"
	"github.com/wujie1993/waves/pkg/setting"
	"github.com/wujie1993/waves/pkg/version"
	"github.com/wujie1993/waves/routers/api/v1"
	"github.com/wujie1993/waves/routers/api/v2"
)

// InitRouter initialize routing information
func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	baseRouter := r.Group(setting.AppSetting.PrefixUrl)
	baseRouter.Static("/web", "./web")
	baseRouter.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查接口
	baseRouter.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, map[string]string{"Health": "ok"})
	})

	// 版本显示接口
	baseRouter.GET("/version", func(c *gin.Context) {
		c.JSON(200, map[string]interface{}{
			"Version": version.Version,
			"Commit":  version.Commit,
			"Build":   version.Build,
			"Author":  version.Author,
			"Golang": map[string]string{
				"Version": version.GoVersion,
			},
		})
	})

	apiRouter := baseRouter.Group("/api")
	// 实体对象接口组
	apiv1 := apiRouter.Group("/v1")
	// apiv1.Use(jwt.JWT())
	{
		ns := apiv1.Group("/namespaces/:namespace")
		{
			app := ns.Group("/apps")
			{
				c := v1.NewAppController()
				app.GET("", c.GetApps)
				app.POST("", c.PostApp)
				app.GET(":name", c.GetApp)
				app.PUT(":name", c.PutApp)
				app.DELETE(":name", c.DeleteApp)
			}

			appInstance := ns.Group("/appinstances")
			{
				c := v1.NewAppInstanceController()
				appInstance.GET("", c.GetAppInstances)
				appInstance.POST("", c.PostAppInstance)
				appInstance.GET(":name", c.GetAppInstance)
				appInstance.PUT(":name", c.PutAppInstance)
				appInstance.DELETE(":name", c.DeleteAppInstance)
			}

			configMap := ns.Group("/configmaps")
			{
				c := v1.NewConfigMapController()
				configMap.GET("", c.GetConfigMaps)
				configMap.POST("", c.PostConfigMap)
				configMap.GET(":name", c.GetConfigMap)
				configMap.PUT(":name", c.PutConfigMap)
				configMap.DELETE(":name", c.DeleteConfigMap)
			}
		}

		k8sconfig := ns.Group("/k8sconfig")
		{
			c := v1.NewK8sConfigController()
			k8sconfig.GET("", c.GetK8ClusterConfigs)
			k8sconfig.POST("", c.PostK8ClusterConfig)
			k8sconfig.GET(":name", c.GetK8ClusterConfig)
			k8sconfig.PUT(":name", c.PutK8ClusterConfig)
			k8sconfig.DELETE(":name", c.DeleteK8sClusterConfig)
		}

		job := apiv1.Group("/jobs")
		{
			c := v1.NewJobController()
			job.GET("", c.GetJobs)
			job.POST("", c.PostJob)
			job.GET(":name", c.GetJob)
			job.PUT(":name", c.PutJob)
			job.DELETE(":name", c.DeleteJob)
			job.GET(":name/log", c.GetJobLog)
		}

		gpu := apiv1.Group("/gpus")
		{
			c := v1.NewGPUController()
			gpu.GET("", c.GetGPUs)
			gpu.POST("", c.PostGPU)
			gpu.GET(":name", c.GetGPU)
			gpu.PUT(":name", c.PutGPU)
			gpu.DELETE(":name", c.DeleteGPU)
		}

		pkg := apiv1.Group("/pkgs")
		{
			c := v1.NewPkgController()
			pkg.GET("", c.GetPkgs)
			pkg.POST("", c.PostPkg)
			pkg.GET(":name", c.GetPkg)
			pkg.PUT(":name", c.PutPkg)
			pkg.DELETE(":name", c.DeletePkg)
		}

		audit := apiv1.Group("/audits")
		{
			c := v1.NewAuditController()
			audit.GET("", c.GetAudits)
			audit.POST("", c.PostAudit)
			audit.GET(":name", c.GetAudit)
			audit.PUT(":name", c.PutAudit)
			audit.DELETE(":name", c.DeleteAudit)
		}

		event := apiv1.Group("/events")
		{
			c := v1.NewEventController()
			event.GET("", c.GetEvents)
			event.POST("", c.PostEvent)
			event.GET(":name", c.GetEvent)
			event.PUT(":name", c.PutEvent)
			event.DELETE(":name", c.DeleteEvent)
		}

		host := apiv1.Group("/hosts")
		{
			c := v1.NewHostController()
			host.GET("", c.GetHosts)
			host.POST("", c.PostHost)
			host.GET(":name", c.GetHost)
			host.PUT(":name", c.PutHost)
			host.DELETE(":name", c.DeleteHost)
		}

		project := apiv1.Group("/project")
		{
			c := v1.NewProjectController()
			project.GET("", c.GetProjects)
			project.GET(":name", c.GetProject)
			project.POST("", c.PostProject)
			project.PUT(":name", c.PutProject)
			project.DELETE(":name", c.DeleteProject)
		}
	}
	apiv2 := apiRouter.Group("/v2")
	{
		ns := apiv2.Group("/namespaces/:namespace")
		{
			appInstance := ns.Group("/appinstances")
			{
				c := v2.NewAppInstanceController()
				appInstance.GET("", c.GetAppInstances)
				appInstance.POST("", c.PostAppInstance)
				appInstance.GET(":name", c.GetAppInstance)
				appInstance.PUT(":name", c.PutAppInstance)
				appInstance.DELETE(":name", c.DeleteAppInstance)
			}
		}

		job := apiv2.Group("/jobs")
		{
			c := v2.NewJobController()
			job.GET("", c.GetJobs)
			job.POST("", c.PostJob)
			job.GET(":name", c.GetJob)
			job.PUT(":name", c.PutJob)
			job.DELETE(":name", c.DeleteJob)
			job.GET(":name/log", c.GetJobLog)
		}
	}

	return r
}
