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

// scan 扫描本地目录上的文件，并提取返回可用于生成代码的模板信息
func scan(pkgPath string) (tpls.CodeTpl, error) {
	// 初始化代码模板
	codeTpl := tpls.CodeTpl{
		Package:    filepath.Base(pkgPath),
		ApiObjects: []string{},
		Registries: []tpls.RegistryTpl{},
	}

	fs := token.NewFileSet()
	f, err := parser.ParseDir(fs, pkgPath, func(f os.FileInfo) bool {
		return true
	}, parser.AllErrors|parser.ParseComments)
	if err != nil {
		log.Error(err)
		return codeTpl, err
	}

	for pkgName, pkgs := range f {
		//log.Debugf(pkgName)
		for fileName, files := range pkgs.Files {
			//log.Debugf(fileName)
			if strings.HasPrefix(filepath.Base(fileName), GENERATED_CODE_FILENAME_PREFIX) || strings.HasSuffix(fileName, "_test.go") {
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
													registryTpl := tpls.RegistryTpl{
														Name:     strings.TrimSuffix(typeSpec.Name.Name, "Registry"),
														Registry: typeSpec.Name.Name,
													}
													log.Tracef("package:%s filename:%s struct:%s is implement base on registry.Registry", pkgName, fileName, typeSpec.Name)
													if genDecl.Doc != nil {
														for _, comment := range genDecl.Doc.List {
															switch comment.Text {
															case "// +namespaced=true":
																registryTpl.Namespaced = true
															}
														}
													}
													codeTpl.Registries = append(codeTpl.Registries, registryTpl)
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
	}

	if len(codeTpl.Registries) > 0 {
		sort.Sort(tpls.SortRegistry(codeTpl.Registries))
	}

	return codeTpl, nil
}
