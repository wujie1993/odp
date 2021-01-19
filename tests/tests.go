package tests

import (
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/setting"
	"github.com/wujie1993/waves/routers"
)

var (
	ServiceEndpoint string
	EtcdEndpoint    string
)

func init() {
	if ServiceEndpoint == "" {
		ServiceEndpoint = "http://localhost:8000/"
	}
	if EtcdEndpoint == "" {
		EtcdEndpoint = "localhost:2379"
	}

	initLog()
	initDB()
}

func initLog() {
	log.SetOutput(os.Stdout)
	log.SetLevel(4)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
}

func initDB() {
	setting.EtcdSetting = &setting.Etcd{
		Endpoints: []string{EtcdEndpoint},
	}
	db.InitKV()
}

func ServeHTTP() {
	routersInit := routers.InitRouter()
	server := &http.Server{
		Addr:    ":8000",
		Handler: routersInit,
	}
	go server.ListenAndServe()
	// wait until the server is ready
	for {
		time.Sleep(time.Second)
		resp, err := http.Get(ServiceEndpoint + "/healthz")
		if err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			continue
		}
		break
	}
}
