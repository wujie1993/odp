package codegen

import (
	"path/filepath"

	//"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/codegen/tpls"
)

const (
	GENERATED_CODE_FILENAME_PREFIX = "zz_generated"
)

type GenOrmOptions struct {
	PkgPath string
}

func GenOrm(opts GenOrmOptions) {
	if err := genByPkgPath(opts.PkgPath); err != nil {
		log.Error(err)
	}
}

func genByPkgPath(pkgPath string) error {
	log.Debugf("Generate orm code to target directory %s", pkgPath)

	codeTpl, err := scan(pkgPath)
	if err != nil {
		return err
	}

	if len(codeTpl.ApiObjects) > 0 {
		// 代码文件生成
		tplMap := map[string]string{
			"encode":   tpls.ENCODE_CODE_TPL,
			"hash":     tpls.HASH_CODE_TPL,
			"deepcopy": tpls.DEEP_COPY_CODE_TPL,
		}
		for name, tpl := range tplMap {
			if err := tpls.RenderCodeFile(codeTpl, tpl, filepath.Join(pkgPath, GENERATED_CODE_FILENAME_PREFIX+"."+name+".go")); err != nil {
				log.Error(err)
				return err
			}
		}
	}

	if len(codeTpl.Registries) > 0 || len(codeTpl.ApiObjects) > 0 {
		if err := tpls.RenderCodeFile(codeTpl, tpls.HELPER_CODE_TPL, filepath.Join(pkgPath, GENERATED_CODE_FILENAME_PREFIX+".helper.go")); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}
