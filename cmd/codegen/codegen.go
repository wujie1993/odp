package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/codegen/cmd"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
}

func main() {
	cmd.Execute()
}
