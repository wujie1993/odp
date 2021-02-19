# 开发文档

## 项目结构

```
├── cmd --- 存放命令行工具的编译入口
│   ├── codegen --- 代码生成工具
│   └── dpctl --- 命令行管理工具
├── conf --- 存放项目所需的配置文件
│   ├── tpls --- 生成ansible脚本相关的模板文件
│   ├── app.ini.sample --- 部署平台配置文件示例
│   └── gpu_types.yml --- gpu型号映射配置(算法示例专用)
├── docker-compose.yml --- docker-compose编排文件
├── Dockerfile --- 容器镜像构建
├── docs --- 存放各类型文档
│   ├── Develop.md --- 开发文档
│   ├── docs.go --- 实现接口文档的生成
│   ├── swagger.json --- json格式swagger接口文档(自动生成)
│   └── swagger.yaml --- yaml格式swagger接口文档(自动生成)
├── go.mod --- 项目包依赖
├── go.sum --- 项目包依赖(自动生成)
├── img_build.sh --- 镜像构建脚本
├── LICENSE --- 项目授权协议
├── main.go --- 程序入口
├── Makefile --- 封装项目编译, 测试等相关快速指令
├── pkg
│   ├── ansible
│   │   ├── args.go --- 填充ansible任务参数的方法实现
│   │   └── tpl.go --- 定义生成ansible任务所需的参数模板
│   ├── client
│   │   ├── rest --- HTTP接口的REST访问库
│   │   ├── v1 --- 封装v1版本资源结构的ClientSet访问库
│   │   ├── v2 --- 封装v1版本资源结构的ClientSet访问库
│   │   └── client.go --- 通过REST访问库封装的ClientSet访问库
│   ├── codegen --- 代码生成器的功能实现
│   │   ├── cmd --- 命令行方法实现
│   │   ├── tpls --- 代码生成模板
│   │   ├── client.go --- 客户端库代码生成
│   │   ├── orm.go --- orm库代码生成
│   │   └── scan.go --- 代码扫描
│   ├── controller --- 基础 HTTP 路由控制器定义
│   ├── db
│   │   └── kv.go --- etcd键值数据库读写封装
│   ├── dpctl --- 命令行管理工具的功能实现
│   │   ├── cmd --- 命令行方法实现
│   │   ├── loader --- 加载本地资源文件
│   │   ├── app.go --- 实现应用管理
│   │   ├── appInstance.go --- 实现应用实例管理
│   │   ├── client.go --- 所有资源管理的调用入口
│   │   ├── common.go --- 公共方法
│   │   ├── configmap.go --- 实现应用实例管理
│   │   ├── host.go --- 实现主机管理
│   │   └── hostPlugin.go --- 实现主机插件管理
│   ├── e
│   │   ├── code.go --- 错误码定义
│   │   └── msg.go --- 错误消息定义
│   ├── file
│   │   └── file.go --- 日志文件记录(未使用)
│   ├── loader --- 实现应用加载
│   ├── operators --- 实现各个实体对象的管理器
│   │   ├── app.go --- 应用管理器
│   │   ├── app_instance.go --- 应用实例管理器
│   │   ├── configmap.go --- 配置字典管理器
│   │   ├── event.go --- 事件管理器
│   │   ├── healthcheck.go --- 应用实例的健康检查项定义
│   │   ├── host.go --- 主机管理器
│   │   ├── job.go --- 任务管理器
│   │   ├── k8s_install.go --- k8s管理器
│   │   ├── operator.go --- 基础管理器定义
│   │   └── README.md
│   ├── orm --- 实现实体对象的数据库读写
│   │   ├── core
│   │   │   ├── api_object.go --- 实体对象接口定义
│   │   │   ├── base_object.go --- 通用实体对象定义
│   │   │   ├── common.go --- 实体对象所用到的一些公共常量与方法
│   │   │   ├── conversion.go --- 结构转换相关的定义
│   │   │   ├── meta.go --- 实体对象元数据定义
│   │   │   ├── option.go --- 存储器操作选项定义
│   │   │   ├── sort.go --- 实体对象排序方法实现
│   │   │   └── status.go --- 实体对象状态定义
│   │   ├── registry
│   │   │   ├── registry.go --- 存储器接口与基础存储器定义
│   │   │   ├── revision.go --- 修订版本记录器定义
│   │   │   └── schema.go --- 数据转换相关方法定义
│   │   ├── conversion.go --- 实体对象跨版本转换方法(未实现)
│   │   ├── runtime --- 运行时资源结构定义
│   │   │   ├── types.go --- 资源结构定义
│   │   │   ├── zz_generated.deepcopy.go --- 资源的内容深度复制方法定义(自动生成)
│   │   │   ├── zz_generated.encode.go --- 资源的序列化方法定义(自动生成)
│   │   │   ├── zz_generated.hash.go --- 资源的哈希计算方法定义(自动生成)
│   │   │   └── zz_generated.helper.go --- 存储器管理入口(自动生成)
│   │   ├── v1 --- v1版本资源和相关方法定义
│   │   │   ├── conversion.go --- v1资源与运行时资源互相转换方法定义
│   │   │   ├── registries.go --- 资源存储器定义
│   │   │   ├── revisions.go --- 资源修订历史记录器定义
│   │   │   ├── types.go --- 资源结构定义
│   │   │   ├── zz_generated.deepcopy.go --- 资源的内容深度复制方法定义(自动生成)
│   │   │   ├── zz_generated.encode.go --- 资源的序列化方法定义(自动生成)
│   │   │   ├── zz_generated.hash.go --- 资源的哈希计算方法定义(自动生成)
│   │   │   └── zz_generated.helper.go --- 存储器管理入口(自动生成)
│   │   ├── v2 --- v2版本资源和相关方法定义
│   │   │   ├── conversion.go --- v1资源与运行时资源互相转换方法定义
│   │   │   ├── registries.go --- 资源存储器定义
│   │   │   ├── revisions.go --- 资源修订历史记录器定义
│   │   │   ├── types.go --- 资源结构定义
│   │   │   ├── zz_generated.deepcopy.go --- 资源的内容深度复制方法定义(自动生成)
│   │   │   ├── zz_generated.encode.go --- 资源的序列化方法定义(自动生成)
│   │   │   ├── zz_generated.hash.go --- 资源的哈希计算方法定义(自动生成)
│   │   │   └── zz_generated.helper.go --- 存储器管理入口(自动生成)
│   │   ├── conversion.go --- 结构版本转换方法入口
│   │   ├── helper.go --- 所有版本资源存储器入口
│   │   └── registry.go --- 数据迁移与初始化方法定义
│   ├── schedule
│   │   ├── scheduler.go --- 实现任务调度器
│   │   └── worker.go --- 实现任务工作器
│   ├── setting
│   │   └── setting.go --- 配置文件加载
│   └── util --- 其他的工具方法
│       ├── md5.go --- md5计算
│       ├── remote_ssh.go --- 通过ssh向远程主机执行指令
│       ├── tailf.go --- 日志文件侦听
│       └── topology.go --- 拓扑结构生成方法定义
├── README.md
├── README_ZH.md
├── routers --- 路由注册与接口实现
│   ├── api --- 接口实现
│   │   └── v1 --- v1版本接口实现
│   │       ├── app.go --- 应用接口实现
│   │       ├── app_instance.go --- 应用实例接口实现
│   │       ├── audit.go --- 审计日志接口实现
│   │       ├── configmap.go --- 配置字典接口实现
│   │       ├── controller.go --- 通用CRUD接口封装, 响应结构体封装以及审计日志记录
│   │       ├── event.go --- 事件日志接口实现
│   │       ├── host.go --- 主机接口实现
│   │       ├── job.go --- 任务接口实现
│   │       └── k8s_config.go --- k8s集群接口实现
│   ├── README.md
│   └── router.go --- 路由注册
├── scripts --- 打包以及部署脚本
├── service --- 部署包相关业务逻辑
├── tests --- 部分测试用例
└── web --- 存放Web前端静态文件
```

## 开发指引

- [定义数据库实体对象](../pkg/orm/README.md)
- [将实体对象映射到接口层](../routers/README.md)
- [为实体对象创建对应的管理器](../pkg/operators/README.md#添加自定义管理器)
