package routers

import (
	"github.com/gin-gonic/gin"

	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	_ "github.com/wujie1993/waves/docs"
	"github.com/wujie1993/waves/pkg/setting"
	"github.com/wujie1993/waves/pkg/version"
	extV1 "github.com/wujie1993/waves/routers/api/extensions/v1"
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

	apiRouter := baseRouter.Group("/api")
	// 实体对象接口组
	apiV1 := apiRouter.Group("/v1")
	// apiV1.Use(jwt.JWT())
	{
		nsCtl := v1.NewNamespaceController()
		apiV1.GET("/namespaces", nsCtl.GetNamespaces)
		apiV1.POST("/namespaces", nsCtl.PostNamespace)

		ns := apiV1.Group("/namespaces/:namespace")
		{
			ns.GET("", nsCtl.GetNamespace)
			ns.PUT("", nsCtl.PutNamespace)
			ns.DELETE("", nsCtl.DeleteNamespace)

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

			k8sconfig := ns.Group("/k8sconfig")
			{
				c := v1.NewK8sConfigController()
				k8sconfig.GET("", c.GetK8ClusterConfigs)
				k8sconfig.POST("", c.PostK8ClusterConfig)
				k8sconfig.GET(":name", c.GetK8ClusterConfig)
				k8sconfig.PUT(":name", c.PutK8ClusterConfig)
				k8sconfig.DELETE(":name", c.DeleteK8sClusterConfig)
			}
		}

		job := apiV1.Group("/jobs")
		{
			c := v1.NewJobController()
			job.GET("", c.GetJobs)
			job.POST("", c.PostJob)
			job.GET(":name", c.GetJob)
			job.PUT(":name", c.PutJob)
			job.DELETE(":name", c.DeleteJob)
			job.GET(":name/log", c.GetJobLog)
		}

		gpu := apiV1.Group("/gpus")
		{
			c := v1.NewGPUController()
			gpu.GET("", c.GetGPUs)
			gpu.POST("", c.PostGPU)
			gpu.GET(":name", c.GetGPU)
			gpu.PUT(":name", c.PutGPU)
			gpu.DELETE(":name", c.DeleteGPU)
		}

		pkg := apiV1.Group("/pkgs")
		{
			c := v1.NewPkgController()
			pkg.GET("", c.GetPkgs)
			pkg.POST("", c.PostPkg)
			pkg.GET(":name", c.GetPkg)
			pkg.PUT(":name", c.PutPkg)
			pkg.DELETE(":name", c.DeletePkg)
		}

		audit := apiV1.Group("/audits")
		{
			c := v1.NewAuditController()
			audit.GET("", c.GetAudits)
			audit.POST("", c.PostAudit)
			audit.GET(":name", c.GetAudit)
			audit.PUT(":name", c.PutAudit)
			audit.DELETE(":name", c.DeleteAudit)
		}

		event := apiV1.Group("/events")
		{
			c := v1.NewEventController()
			event.GET("", c.GetEvents)
			event.POST("", c.PostEvent)
			event.GET(":name", c.GetEvent)
			event.PUT(":name", c.PutEvent)
			event.DELETE(":name", c.DeleteEvent)
		}

		host := apiV1.Group("/hosts")
		{
			c := v1.NewHostController()
			host.GET("", c.GetHosts)
			host.POST("", c.PostHost)
			host.GET(":name", c.GetHost)
			host.PUT(":name", c.PutHost)
			host.DELETE(":name", c.DeleteHost)
		}

		project := apiV1.Group("/project")
		{
			c := v1.NewProjectController()
			project.GET("", c.GetProjects)
			project.GET(":name", c.GetProject)
			project.POST("", c.PostProject)
			project.PUT(":name", c.PutProject)
			project.DELETE(":name", c.DeleteProject)
		}

		topology := apiV1.Group("/topology")
		{
			c := extV1.NewTopologyController()
			topology.GET("", c.GetTopology)
		}

		revision := apiV1.Group("/revisions")
		{
			c := v1.NewRevisionController()
			revision.GET("", c.GetRevisions)
			revision.POST("", c.PostRevision)
			revision.GET(":name", c.GetRevision)
			revision.PUT(":name", c.PutRevision)
			revision.DELETE(":name", c.DeleteRevision)
		}
	}
	apiV2 := apiRouter.Group("/v2")
	{
		ns := apiV2.Group("/namespaces/:namespace")
		{
			appInstance := ns.Group("/appinstances")
			{
				c := v2.NewAppInstanceController()
				appInstance.GET("", c.GetAppInstances)
				appInstance.POST("", c.PostAppInstance)
				appInstance.GET(":name", c.GetAppInstance)
				appInstance.PUT(":name", c.PutAppInstance)
				appInstance.DELETE(":name", c.DeleteAppInstance)
				appInstance.GET(":name/revisions", c.GetAppInstanceRevisions)
				appInstance.GET(":name/revisions/:revision", c.GetAppInstanceRevision)
				appInstance.PUT(":name/revisions/:revision", c.PutAppInstanceRevision)
				appInstance.DELETE(":name/revisions/:revision", c.DeleteAppInstanceRevision)
			}
		}

		job := apiV2.Group("/jobs")
		{
			c := v2.NewJobController()
			job.GET("", c.GetJobs)
			job.POST("", c.PostJob)
			job.GET(":name", c.GetJob)
			job.PUT(":name", c.PutJob)
			job.DELETE(":name", c.DeleteJob)
			job.GET(":name/log", c.GetJobLog)
		}

		host := apiV2.Group("/hosts")
		{
			c := v2.NewHostController()
			host.GET("", c.GetHosts)
			host.POST("", c.PostHost)
			host.GET(":name", c.GetHost)
			host.PUT(":name", c.PutHost)
			host.DELETE(":name", c.DeleteHost)
		}
	}

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

	// 健康检查接口
	baseRouter.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, map[string]string{"Health": "ok"})
	})

	return r
}
