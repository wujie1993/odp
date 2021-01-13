package codegen

import (
	"path/filepath"

	//"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/codegen/tpls"
)

type GenClientOptions struct {
	InputPkgPath  string
	OutputPkgPath string
}

func GenClient(opts GenClientOptions) {
	codeTpl, err := scan(opts.InputPkgPath)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("Generate client code to target directory %s", opts.OutputPkgPath)

	if err := tpls.RenderCodeFile(codeTpl, tpls.CLIENT_CODE_TPL, filepath.Join(opts.OutputPkgPath, GENERATED_CODE_FILENAME_PREFIX+".client.go")); err != nil {
		log.Error(err)
		return
	}
}
