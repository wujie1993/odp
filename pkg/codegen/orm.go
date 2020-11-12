package codegen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	//"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/codegen/tpls"
)

const (
	GENERATED_CODE_FILENAME_PREFIX = "zz_generated"
)

type SortByName []string

func (s SortByName) Len() int {
	return len(s)
}

func (s SortByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortByName) Less(i, j int) bool {
	return strings.Compare(s[i], s[j]) < 0
}

type GenOrmOptions struct {
	PkgPath string
}

func GenOrm(opts GenOrmOptions) {
	if err := genByPkgPath(opts.PkgPath); err != nil {
		log.Error(err)
	}
}

func genByPkgPath(pkgPath string) error {
	log.Debugf("generate orm codes from path %s", pkgPath)

	fs := token.NewFileSet()
	f, err := parser.ParseDir(fs, pkgPath, func(f os.FileInfo) bool {
		return true
	}, parser.AllErrors)
	if err != nil {
		log.Error(err)
		return err
	}

	// 初始化模板结构
	codeTpl := tpls.CodeTpl{
		Package:    filepath.Base(pkgPath),
		ApiObjects: []string{},
		Registries: []tpls.RegistryTpl{},
	}

	for pkgName, pkgs := range f {
		//log.Debugf(pkgName)
		for fileName, files := range pkgs.Files {
			//log.Debugf(fileName)
			if filepath.Base(fileName) == GENERATED_CODE_FILENAME_PREFIX+"json.go" || filepath.Base(fileName) == GENERATED_CODE_FILENAME_PREFIX+"yaml.go" || strings.HasSuffix(fileName, "_test.go") {
				continue
			}
			for _, decl := range files.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok {
					for _, spec := range genDecl.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							//log.Debug(typeSpec.Name)
							if structType, ok := typeSpec.Type.(*ast.StructType); ok {
								for _, field := range structType.Fields.List {
									//log.Debug(field.Names)
									if selectorExpr, ok := field.Type.(*ast.SelectorExpr); ok {
										if selectorExpr.Sel.Name == "BaseApiObj" {
											if x, ok := selectorExpr.X.(*ast.Ident); ok {
												if x.Name == "core" {
													codeTpl.ApiObjects = append(codeTpl.ApiObjects, typeSpec.Name.Name)
													log.Tracef("package:%s filename:%s struct:%s is implement base on core.BaseApiObj", pkgName, fileName, typeSpec.Name)
												}
											}
										}
										if selectorExpr.Sel.Name == "Registry" {
											if x, ok := selectorExpr.X.(*ast.Ident); ok {
												if x.Name == "registry" {
													codeTpl.Registries = append(codeTpl.Registries, tpls.RegistryTpl{Name: strings.TrimSuffix(typeSpec.Name.Name, "Registry"), Registry: typeSpec.Name.Name})
													log.Tracef("package:%s filename:%s struct:%s is implement base on registry.Registry", pkgName, fileName, typeSpec.Name)
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if len(codeTpl.ApiObjects) > 0 {
		// 对象按名称排序
		sort.Sort(SortByName(codeTpl.ApiObjects))

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
		sort.Sort(tpls.SortRegistry(codeTpl.Registries))
		if err := tpls.RenderCodeFile(codeTpl, tpls.HELPER_CODE_TPL, filepath.Join(pkgPath, GENERATED_CODE_FILENAME_PREFIX+".helper.go")); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}
