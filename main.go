package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/loader"
	"github.com/wujie1993/waves/pkg/operators"
	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/schedule"
	"github.com/wujie1993/waves/pkg/setting"
	"github.com/wujie1993/waves/routers"
)

func init() {
	setting.Setup()

	// 设置日志输出
	log.SetOutput(os.Stdout)
	log.SetLevel(setting.AppSetting.LogLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	log.SetReportCaller(true)

	// 初始化数据库连接
	db.InitKV()

	// 初始化底层数据
	orm.InitStorage()

	// 加载应用
	loadApps()
}

func loadApps() {
	// 在服务启动时先同步刷新一遍应用，再以每30秒为间隔异步刷新应用
	loader.LoadApps([]string{core.AppCategoryThirdParty, core.AppCategoryHostPlugin}, filepath.Join(setting.AnsibleSetting.PlaybooksDir, setting.PlaybooksAppsDir, setting.AppsYml))
	loader.LoadPkgs([]string{core.AppCategoryCustomize, core.AppCategoryAlgorithmPlugin}, setting.PackageSetting.PkgPath)
	go func() {
		for {
			time.Sleep(30 * time.Second)
			loader.LoadApps([]string{core.AppCategoryThirdParty, core.AppCategoryHostPlugin}, filepath.Join(setting.AnsibleSetting.PlaybooksDir, setting.PlaybooksAppsDir, setting.AppsYml))
			loader.LoadPkgs([]string{core.AppCategoryCustomize, core.AppCategoryAlgorithmPlugin}, setting.PackageSetting.PkgPath)
		}
	}()
}

func loadPlugins(ctx context.Context) {
	s := schedule.NewScheduler()
	go s.Run(ctx)

	appOperator := operators.NewAppOperator()
	go appOperator.Run(ctx)

	configMapOperator := operators.NewConfigMapOperator()
	go configMapOperator.Run(ctx)

	jobOperator := operators.NewJobOperator()
	go jobOperator.Run(ctx)

	eventOperator := operators.NewEventOperator()
	go eventOperator.Run(ctx)

	hostOperator := operators.NewHostOperator()
	go hostOperator.Run(ctx)

	k8sOperator := operators.NewK8sInstallOperator()
	go k8sOperator.Run(ctx)

	appInstanceOperator := operators.NewAppInstanceOperator()
	go appInstanceOperator.Run(ctx)
}

// @title Golang Gin API
// @version 1.0
// @description An example of gin
// @termsOfService https://github.com/wujie1993/waves
// @license.name MIT
// @license.url https://github.com/wujie1993/waves/blob/master/LICENSE

// @BasePath /deployer

// @tag.name App
// @tag.description 应用

// @tag.name AppInstance
// @tag.description 应用实例

// @tag.name Host
// @tag.description 主机

// @tag.name Job
// @tag.description 任务

// @tag.name ConfigMap
// @tag.description 配置字典

// @tag.name K8sConfig
// @tag.description K8s集群

// @tag.name Pkg
// @tag.description 部署包

// @tag.name Audit
// @tag.description 审计

// @tag.name Event
// @tag.description 事件

// @tag.name GPU
// @tag.description 显卡

// @tag.name Namespace
// @tag.description 命名空间

// @tag.name Project
// @tag.description 项目

// @tag.name Revision
// @tag.description 修订历史

// @tag.name Topology
// @tag.description 拓扑

func main() {
	flag.Parse()

	ctx := context.Background()
	loadPlugins(ctx)

	gin.SetMode(setting.ServerSetting.RunMode)

	routersInit := routers.InitRouter()
	// readTimeout := setting.ServerSetting.ReadTimeout
	// writeTimeout := setting.ServerSetting.WriteTimeout
	endPoint := fmt.Sprintf(":%d", setting.ServerSetting.HttpPort)
	// maxHeaderBytes := 1 << 20

	server := &http.Server{
		Addr:    endPoint,
		Handler: routersInit,
		// ReadTimeout:    readTimeout,
		// WriteTimeout:   writeTimeout,
		// MaxHeaderBytes: maxHeaderBytes,
	}

	log.Printf("[info] start http server listening %s", endPoint)

	quit := make(chan os.Signal)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Warn(err)
			close(quit)
		}
	}()

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit
	switch sig {
	case os.Interrupt:
		log.Warnf("received interrupt signal")
	case syscall.SIGTERM:
		log.Warnf("received terminal signal")
	}

	log.Warnf("shutting down server ...")
	shutdownCtx, _ := context.WithTimeout(ctx, 5*time.Second)
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error(err)
	}
	log.Warnf("server exited")

	// If you want Graceful Restart, you need a Unix system and download github.com/fvbock/endless
	//endless.DefaultReadTimeOut = readTimeout
	//endless.DefaultWriteTimeOut = writeTimeout
	//endless.DefaultMaxHeaderBytes = maxHeaderBytes
	//server := endless.NewServer(endPoint, routersInit)
	//server.BeforeBegin = func(add string) {
	//	log.Printf("Actual pid is %d", syscall.Getpid())
	//}
	//
	//err := server.ListenAndServe()
	//if err != nil {
	//	log.Printf("Server err: %v", err)
	//}
}
