package extv1

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/controller"
	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/orm/v2"
	"github.com/wujie1993/waves/pkg/util"
)

const (
	TopologyFormatCSV     = "csv"
	TopologyFormatMermaid = "mermaid"
)

type TopologyController struct {
	controller.BaseController
}

// @summary 导出拓扑结构
// @tags Topology
// @produce json
// @accept json
// @param resourceKind query string true "资源类别" Enums(appInstance, host)
// @param resourceNamespace query string true "资源命名空间" default(default)
// @param resourceName query string false "资源名称"
// @param format query string false "导出格式" Enums(mermaid, csv)
// @param download query boolean false "下载文件"
// @success 200 {object} controller.Response
// @failure 500 {object} controller.Response
// @router /api/v1/topology [get]
func (c *TopologyController) GetTopology(ctx *gin.Context) {
	resourceKind := ctx.Query("resourceKind")
	resourceNamespace := ctx.Query("resourceNamespace")
	resourceName := ctx.Query("resourceName")
	format := ctx.Query("format")
	download := ctx.Query("download")

	log.Debug(resourceKind, resourceNamespace, resourceName, format, download)
	result, err := getTopology(resourceKind, resourceNamespace, resourceName, format)
	if err != nil {
		c.Response(ctx, 500, e.ERROR, err.Error(), nil)
		return
	}

	if download == "true" {
		var filename string
		if resourceName == "" {
			filename = fmt.Sprintf("%s.%s", resourceKind, format)
		} else {
			filename = fmt.Sprintf("%s.%s", resourceName, format)
		}
		ctx.Header("Content-Disposition", "attachment; filename="+filename)
		ctx.Header("Content-Type", "application/octet-stream; charset=utf-8")
		ctx.Data(200, "text/"+format, []byte(result))
	} else {
		c.Response(ctx, 200, e.SUCCESS, "", result)
	}
}

func getTopology(kind, namespace, name, format string) (string, error) {
	switch kind {
	case core.KindHost:
		return getHostTopology(name, format)
	case core.KindAppInstance:
		return getAppInstanceTopology(namespace, name, format)
	}
	return "", nil
}

func getAppInstanceTopology(namespace, name, format string) (string, error) {
	switch format {
	case TopologyFormatMermaid:
		return getAppInstanceMermaidTopology(namespace, name)
	case TopologyFormatCSV:
		return getAppInstanceCSVTopology(namespace, name)
	}
	return "", nil
}

func getAppInstanceMermaidTopology(namespace, name string) (string, error) {
	helper := orm.GetHelper()

	appInstanceObj, err := helper.V2.AppInstance.Get(context.TODO(), namespace, name)
	if err != nil {
		return "", err
	}
	appInstance := appInstanceObj.(*v2.AppInstance)

	// 获取关联的应用
	appObj, err := helper.V1.App.Get(context.TODO(), core.DefaultNamespace, appInstance.Spec.AppRef.Name)
	if err != nil {
		log.Error(err)
		return "", err
	}
	if appObj == nil {
		return "", nil
	}
	app := appObj.(*v1.App)

	versionApp, ok := app.GetVersion(appInstance.Spec.AppRef.Version)
	if !ok {
		return "", nil
	}

	md := util.MermaidFlow{
		Orientation: util.MermaidFlowOrientationLR,
		Nodes:       map[string]util.MermaidFlowNode{},
		Links:       []util.MermaidFlowLink{},
	}

	// 添加一级应用实例根节点
	instanceDesc, ok := appInstance.Metadata.Annotations["ShortName"]
	if !ok {
		instanceDesc = appInstance.Metadata.Name
	}
	md.Nodes[appInstance.Metadata.Name] = util.MermaidFlowNode{
		Desc:  fmt.Sprintf("应用实例 %s (%s)", appInstance.Metadata.Name, instanceDesc),
		Shape: util.MermaidFlowShapeStadium,
	}

	for _, module := range appInstance.Spec.Modules {
		appModule, ok := versionApp.GetModule(module.Name)
		if !ok {
			continue
		}
		moduleDesc := appModule.Desc
		if moduleDesc == "" {
			moduleDesc = appModule.Name
		}

		// 添加二级模块节点
		moduleDesc = fmt.Sprintf("模块 %s (%s) %s", appModule.Name, moduleDesc, module.AppVersion)
		md.Nodes[appModule.Name] = util.MermaidFlowNode{
			Desc:  moduleDesc,
			Shape: util.MermaidFlowShapeRoundEdges,
		}

		// 添加一级实例实例到二级模块的连接
		md.Links = append(md.Links, util.MermaidFlowLink{
			Src: util.MermaidFlowLinkEndpoint{
				Name: appInstance.Metadata.Name,
			},
			Dst: util.MermaidFlowLinkEndpoint{
				Name:  appModule.Name,
				Arrow: util.MermaidFlowArrowTriangle,
			},
			LinkType: util.MermaidFlowLinkTypeNormal,
		})

		for _, replica := range module.Replicas {
			for _, hostRef := range replica.HostRefs {
				// 获取插件关联的主机
				hostObj, err := helper.V1.Host.Get(context.TODO(), core.DefaultNamespace, hostRef)
				if err != nil {
					log.Error(err)
					return "", err
				}
				if hostObj == nil {
					err := e.Errorf("host %s not found", hostRef)
					log.Error(err)
					return "", err
				}
				host := hostObj.(*v1.Host)

				hostDesc, ok := host.Metadata.Annotations["ShortName"]
				if !ok {
					hostDesc = host.Metadata.Name
				}
				// 添加三级主机节点
				md.Nodes[hostRef] = util.MermaidFlowNode{
					Desc:  fmt.Sprintf("主机 %s (%s)", host.Metadata.Name, hostDesc),
					Shape: util.MermaidFlowShapeRectangle,
				}

				// 添加二级模块到三级主机的连接
				md.Links = append(md.Links, util.MermaidFlowLink{
					Src: util.MermaidFlowLinkEndpoint{
						Name: module.Name,
					},
					Dst: util.MermaidFlowLinkEndpoint{
						Name:  host.Metadata.Name,
						Arrow: util.MermaidFlowArrowTriangle,
					},
					LinkType: util.MermaidFlowLinkTypeNormal,
				})
			}
		}
	}
	ret, _ := util.RenderMermaidFlowChart(md)
	return ret, nil
}

func getAppInstanceCSVTopology(namespace, name string) (string, error) {
	if name == "" {
		return getAllHostsCSVTopology()
	}

	helper := orm.GetHelper()

	appInstanceObj, err := helper.V2.AppInstance.Get(context.TODO(), namespace, name)
	if err != nil {
		return "", err
	}
	appInstance := appInstanceObj.(*v2.AppInstance)

	// 获取关联的应用
	appObj, err := helper.V1.App.Get(context.TODO(), core.DefaultNamespace, appInstance.Spec.AppRef.Name)
	if err != nil {
		log.Error(err)
		return "", err
	}
	if appObj == nil {
		return "", nil
	}
	app := appObj.(*v1.App)

	versionApp, ok := app.GetVersion(appInstance.Spec.AppRef.Version)
	if !ok {
		return "", nil
	}

	// 初始化csv表格
	csvWriter := util.NewCSVWriter([]string{
		"应用实例",
		"类别",
		"应用名",
		"应用版本",
		"模块名",
		"模块版本",
		"切片序号",
		"所在主机",
	})

	// 获取应用实例描述
	instanceDesc, ok := appInstance.Metadata.Annotations["ShortName"]
	if !ok {
		instanceDesc = appInstance.Metadata.Name
	}
	// 获取应用类别
	categoryDesc := getCategoryDesc(appInstance.Spec.Category)

	for _, module := range appInstance.Spec.Modules {
		appModule, ok := versionApp.GetModule(module.Name)
		if !ok {
			continue
		}

		moduleDesc := appModule.Desc
		if moduleDesc == "" {
			moduleDesc = appModule.Name
		}

		moduleVersion := module.AppVersion
		if moduleVersion == "" {
			moduleVersion = versionApp.Version
		}

		for replicaIndex, replica := range module.Replicas {
			for _, hostRef := range replica.HostRefs {
				// 获取插件关联的主机
				hostObj, err := helper.V1.Host.Get(context.TODO(), core.DefaultNamespace, hostRef)
				if err != nil {
					log.Error(err)
					return "", err
				}
				if hostObj == nil {
					err := e.Errorf("host %s not found", hostRef)
					log.Error(err)
					return "", err
				}
				host := hostObj.(*v1.Host)

				// 获取主机描述
				hostDesc, ok := host.Metadata.Annotations["ShortName"]
				if !ok {
					hostDesc = host.Metadata.Name
				}

				// 添加csv记录
				csvWriter.WriteRow([]string{
					fmt.Sprintf("%s(%s)", appInstance.Metadata.Name, instanceDesc),
					categoryDesc,
					fmt.Sprintf("%s(%s)", app.Metadata.Name, versionApp.Desc),
					versionApp.Version,
					fmt.Sprintf("%s(%s)", appModule.Name, moduleDesc),
					moduleVersion,
					fmt.Sprint(replicaIndex),
					fmt.Sprintf("%s(%s)", host.Metadata.Name, hostDesc),
				})
			}
		}
	}
	return csvWriter.String(), nil
}

func getHostTopology(name, format string) (string, error) {
	switch format {
	case TopologyFormatMermaid:
		return getHostMermaidTopology(name)
	case TopologyFormatCSV:
		return getHostCSVTopology(name)
	}
	return "", nil
}

func getHostMermaidTopology(name string) (string, error) {
	helper := orm.GetHelper()

	hostObj, err := helper.V1.Host.Get(context.TODO(), "", name)
	if err != nil {
		return "", err
	} else if hostObj == nil {
		return "", nil
	}
	host := hostObj.(*v1.Host)

	// 初始化mermaid结构
	hostSubGraph := util.MermaidFlowSubGraph{
		Nodes: map[string]util.MermaidFlowNode{},
	}
	instanceNodes := map[string]util.MermaidFlowNode{}
	instanceModuleLinks := []util.MermaidFlowLink{}

	appInstanceObjs, err := helper.V2.AppInstance.List(context.TODO(), core.DefaultNamespace)
	if err != nil {
		return "", nil
	}

	appObjs, err := helper.V1.App.List(context.TODO(), core.DefaultNamespace)
	if err != nil {
		return "", nil
	}

	for _, appInstanceObj := range appInstanceObjs {
		appInstance := appInstanceObj.(*v2.AppInstance)
		// 只统计已安装的应用实例
		if appInstance.Status.Phase != core.PhaseInstalled {
			continue
		}
		for _, module := range appInstance.Spec.Modules {
			// 用于标记当前模块是否与目标主机建立关联，避免重复建立关联
			link := false
			moduleNodeName := appInstance.Metadata.Name + "_" + module.Name
			for _, replica := range module.Replicas {
				for _, hostRef := range replica.HostRefs {
					if hostRef == host.Metadata.Name {
						// 添加应用实例节点
						if _, ok := instanceNodes[appInstance.Metadata.Name]; !ok {
							appInstanceDesc, ok := appInstance.Metadata.Annotations["ShortName"]
							if !ok {
								appInstanceDesc = appInstance.Metadata.Name
							}
							instanceNodes[appInstance.Metadata.Name] = util.MermaidFlowNode{
								Desc:  fmt.Sprintf("应用实例 %s (%s)", appInstance.Metadata.Name, appInstanceDesc),
								Shape: util.MermaidFlowShapeStadium,
							}
						}

						// 添加模块节点
						if _, ok := hostSubGraph.Nodes[moduleNodeName]; !ok {
							var moduleDesc string
							for _, appObj := range appObjs {
								app := appObj.(*v1.App)
								if app.Metadata.Name == appInstance.Spec.AppRef.Name {
									appModule, ok := app.GetVersionModule(appInstance.Spec.AppRef.Version, module.Name)
									if !ok {
										break
									}
									moduleDesc = appModule.Desc
									break
								}
							}
							if moduleDesc == "" {
								moduleDesc = module.Name
							}
							hostSubGraph.Nodes[moduleNodeName] = util.MermaidFlowNode{
								Desc:  fmt.Sprintf("模块 %s (%s) %s", module.Name, moduleDesc, module.AppVersion),
								Shape: util.MermaidFlowShapeRoundEdges,
							}
						}

						link = true
						break
					}
				}
				if link {
					break
				}
			}
			if link {
				instanceModuleLinks = append(instanceModuleLinks, util.MermaidFlowLink{
					Src: util.MermaidFlowLinkEndpoint{
						Name: appInstance.Metadata.Name,
					},
					Dst: util.MermaidFlowLinkEndpoint{
						Name:  moduleNodeName,
						Arrow: util.MermaidFlowArrowTriangle,
					},
					LinkType: util.MermaidFlowLinkTypeNormal,
				})
			}
		}
	}
	md := util.MermaidFlow{
		Orientation: util.MermaidFlowOrientationLR,
		Nodes:       instanceNodes,
		Links:       instanceModuleLinks,
		SubGraphs: map[string]util.MermaidFlowSubGraph{
			// 添加主机组
			fmt.Sprintf("主机 %s", host.Spec.SSH.Host): hostSubGraph,
		},
	}
	return util.RenderMermaidFlowChart(md)
}

func getHostCSVTopology(name string) (string, error) {
	if name == "" {
		return getAllHostsCSVTopology()
	}

	helper := orm.GetHelper()

	hostObj, err := helper.V1.Host.Get(context.TODO(), "", name)
	if err != nil {
		return "", err
	} else if hostObj == nil {
		return "", nil
	}
	host := hostObj.(*v1.Host)

	hostDesc, ok := host.Metadata.Annotations["ShortName"]
	if !ok {
		hostDesc = host.Metadata.Name
	}

	appInstanceObjs, err := helper.V2.AppInstance.List(context.TODO(), core.DefaultNamespace)
	if err != nil {
		return "", nil
	}

	appObjs, err := helper.V1.App.List(context.TODO(), core.DefaultNamespace)
	if err != nil {
		return "", nil
	}

	// 初始化csv表格
	csvWriter := util.NewCSVWriter([]string{
		"应用实例",
		"类别",
		"应用名",
		"应用版本",
		"模块名",
		"模块版本",
		"切片序号",
		"所在主机",
	})

	for _, appInstanceObj := range appInstanceObjs {
		appInstance := appInstanceObj.(*v2.AppInstance)
		// 只统计已安装的应用实例
		if appInstance.Status.Phase != core.PhaseInstalled {
			continue
		}
		// 获取应用实例类别
		categoryDesc := getCategoryDesc(appInstance.Spec.Category)

		for _, module := range appInstance.Spec.Modules {
			// 用于标记当前模块是否与目标主机建立关联，避免重复建立关联
			for replicaIndex, replica := range module.Replicas {
				for _, hostRef := range replica.HostRefs {
					if hostRef == host.Metadata.Name {
						appInstanceDesc, ok := appInstance.Metadata.Annotations["ShortName"]
						if !ok {
							appInstanceDesc = appInstance.Metadata.Name
						}

						for _, appObj := range appObjs {
							app := appObj.(*v1.App)
							if app.Metadata.Name == appInstance.Spec.AppRef.Name {
								versionApp, ok := app.GetVersion(appInstance.Spec.AppRef.Version)
								if !ok {
									break
								}

								appModule, ok := versionApp.GetModule(module.Name)
								if !ok {
									break
								}

								moduleDesc := appModule.Desc
								if moduleDesc == "" {
									moduleDesc = module.Name
								}

								moduleVersion := module.AppVersion
								if moduleVersion == "" {
									moduleVersion = versionApp.Version
								}

								// 添加csv记录
								csvWriter.WriteRow([]string{
									fmt.Sprintf("%s(%s)", appInstance.Metadata.Name, appInstanceDesc),
									categoryDesc,
									fmt.Sprintf("%s(%s)", app.Metadata.Name, versionApp.Desc),
									appInstance.Spec.AppRef.Version,
									fmt.Sprintf("%s(%s)", appModule.Name, moduleDesc),
									moduleVersion,
									fmt.Sprint(replicaIndex),
									fmt.Sprintf("%s(%s)", host.Metadata.Name, hostDesc),
								})

								break
							}
						}
					}
				}
			}
		}
	}
	return csvWriter.String(), nil
}

func getAllHostsCSVTopology() (string, error) {
	helper := orm.GetHelper()

	// 初始化csv表格
	csvWriter := util.NewCSVWriter([]string{
		"应用实例",
		"类别",
		"应用名",
		"应用版本",
		"模块名",
		"模块版本",
		"切片序号",
		"所在主机",
	})

	appInstanceObjs, err := helper.V2.AppInstance.List(context.TODO(), core.DefaultNamespace)
	if err != nil {
		return "", nil
	}

	appObjs, err := helper.V1.App.List(context.TODO(), core.DefaultNamespace)
	if err != nil {
		return "", nil
	}

	hostObjs, err := helper.V1.Host.List(context.TODO(), "")
	if err != nil {
		return "", err
	}

	for _, hostObj := range hostObjs {
		host := hostObj.(*v1.Host)
		hostDesc, ok := host.Metadata.Annotations["ShortName"]
		if !ok {
			hostDesc = host.Metadata.Name
		}
		for _, appInstanceObj := range appInstanceObjs {
			appInstance := appInstanceObj.(*v2.AppInstance)
			// 只统计已安装的应用实例
			if appInstance.Status.Phase != core.PhaseInstalled {
				continue
			}
			// 获取应用实例类别
			categoryDesc := getCategoryDesc(appInstance.Spec.Category)

			for _, module := range appInstance.Spec.Modules {
				// 用于标记当前模块是否与目标主机建立关联，避免重复建立关联
				for replicaIndex, replica := range module.Replicas {
					for _, hostRef := range replica.HostRefs {
						if hostRef == host.Metadata.Name {
							appInstanceDesc, ok := appInstance.Metadata.Annotations["ShortName"]
							if !ok {
								appInstanceDesc = appInstance.Metadata.Name
							}

							for _, appObj := range appObjs {
								app := appObj.(*v1.App)
								if app.Metadata.Name == appInstance.Spec.AppRef.Name {
									versionApp, ok := app.GetVersion(appInstance.Spec.AppRef.Version)
									if !ok {
										break
									}

									appModule, ok := versionApp.GetModule(module.Name)
									if !ok {
										break
									}

									moduleDesc := appModule.Desc
									if moduleDesc == "" {
										moduleDesc = module.Name
									}

									moduleVersion := module.AppVersion
									if moduleVersion == "" {
										moduleVersion = versionApp.Version
									}

									// 添加csv记录
									csvWriter.WriteRow([]string{
										fmt.Sprintf("%s(%s)", appInstance.Metadata.Name, appInstanceDesc),
										categoryDesc,
										fmt.Sprintf("%s(%s)", app.Metadata.Name, versionApp.Desc),
										appInstance.Spec.AppRef.Version,
										fmt.Sprintf("%s(%s)", appModule.Name, moduleDesc),
										moduleVersion,
										fmt.Sprint(replicaIndex),
										fmt.Sprintf("%s(%s)", host.Metadata.Name, hostDesc),
									})

									break
								}
							}
						}
					}
				}
			}
		}
	}
	return csvWriter.String(), nil
}

func getCategoryDesc(category string) string {
	switch category {
	case core.AppCategoryCustomize:
		return "业务应用"
	case core.AppCategoryHostPlugin:
		return "主机插件"
	case core.AppCategoryThirdParty:
		return "基础组件"
	}
	return ""
}

func NewTopologyController() TopologyController {
	return TopologyController{
		BaseController: controller.NewController(nil),
	}
}
