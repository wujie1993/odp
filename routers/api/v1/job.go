package v1

import (
	"context"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/controller"
	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/setting"
)

type JobController struct {
	controller.BaseController
}

// @summary 获取所有任务
// @tags Job
// @produce json
// @accept json
// @success 200 {object} controller.Response{Data=[]v1.Job}
// @failure 500 {object} controller.Response
// @router /api/v1/jobs [get]
func (c *JobController) GetJobs(ctx *gin.Context) {
	c.List(ctx)
}

// @summary 获取单个任务
// @tags Job
// @produce json
// @accept json
// @param name path string true "任务名称"
// @success 200 {object} controller.Response{Data=v1.Job}
// @failure 500 {object} controller.Response
// @router /api/v1/jobs/{name} [get]
func (c *JobController) GetJob(ctx *gin.Context) {
	c.Get(ctx)
}

// @summary 获取单个任务的运行日志
// @tags Job
// @produce json
// @accept json
// @param name path string true "任务名称"
// @param watch query boolean false "开启侦听模式"
// @param download query boolean false "下载日志文件"
// @success 200 {object} controller.Response{Data=v1.Job}
// @failure 500 {object} controller.Response
// @router /api/v1/jobs/{name}/log [get]
func (c *JobController) GetJobLog(ctx *gin.Context) {
	name := ctx.Param("name")
	watch := ctx.Query("watch")
	download := ctx.Query("download")
	jobDirs := path.Join(setting.AppSetting.DataDir, setting.JobsDir)

	helper := orm.GetHelper()

	if download == "true" {
		jobPath, err := helper.V1.Job.GetLogPath(jobDirs, name)
		if err != nil {
			c.Response(ctx, 500, e.ERROR, err.Error(), nil)
			return
		} else if jobPath == "" {
			c.Response(ctx, 404, e.ERROR, "job not found", nil)
			return
		}
		ctx.Writer.Header().Add("Content-Disposition", "attachment; filename="+name+".log")
		ctx.Writer.Header().Add("Content-Type", "application/octet-stream")
		ctx.File(jobPath)
	} else if watch == "true" {
		// 侦听模式
		upgrader := websocket.Upgrader{}
		//upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			log.Error(err)
			c.Response(ctx, 500, e.ERROR, err.Error(), nil)
			return
		}
		defer conn.Close()

		watchCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		logWatcher, err := helper.V1.Job.WatchLog(watchCtx, jobDirs, name)
		if err != nil {
			log.Error(err)
			return
		}
		for {
			select {
			case line, ok := <-logWatcher:
				if !ok {
					return
				}
				err = conn.WriteMessage(websocket.TextMessage, []byte(line))
				if err != nil {
					log.Error(err)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	} else {
		// 非侦听模式
		result, err := helper.V1.Job.GetLog(jobDirs, name)
		if err != nil {
			c.Response(ctx, 500, e.ERROR, err.Error(), nil)
			return
		}

		c.Response(ctx, 200, e.SUCCESS, "", string(result))
	}
}

// @summary 创建单个任务
// @tags Job
// @produce json
// @accept json
// @param body body v1.Job true "任务信息"
// @success 200 {object} controller.Response{Data=v1.Job}
// @failure 500 {object} controller.Response
// @router /api/v1/jobs [post]
func (c *JobController) PostJob(ctx *gin.Context) {
	c.Create(ctx)
}

// @summary 更新单个任务
// @tags Job
// @produce json
// @accept json
// @param name path string true "任务名称"
// @param body body v1.Job true "任务信息"
// @success 200 {object} controller.Response{Data=v1.Job}
// @failure 500 {object} controller.Response
// @router /api/v1/jobs/{name} [put]
func (c *JobController) PutJob(ctx *gin.Context) {
	c.Update(ctx)
}

// @summary 删除单个任务
// @tags Job
// @produce json
// @accept json
// @param name path string true "任务名称"
// @success 200 {object} controller.Response{Data=v1.Job}
// @failure 500 {object} controller.Response
// @router /api/v1/jobs/{name} [delete]
func (c *JobController) DeleteJob(ctx *gin.Context) {
	c.Delete(ctx)
}

func NewJobController() JobController {
	return JobController{
		BaseController: controller.NewController(v1.NewJobRegistry()),
	}
}
