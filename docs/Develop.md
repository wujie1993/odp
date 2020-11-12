# 开发文档

## 项目结构

```
├── conf --- 存放项目所需的配置文件
│   └── app.ini --- 部署平台配置文件模板
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
├── middleware --- http服务中间件(未使用)
├── pkg
│   ├── ansible
│   │   └── tpl.go --- 定义生成ansible任务所需的参数模板
│   ├── db
│   │   ├── kv.go --- etcd键值数据库读写封装
│   │   └── kv_test.go
│   ├── e --- 定义http请求中返回的错误信息
│   │   ├── code.go --- 错误码
│   │   └── msg.go --- 错误消息
│   ├── file
│   │   └── file.go --- 日志文件记录(未使用)
│   ├── loader --- 实现应用加载
│   ├── operators --- 实现各个实体对象的管理器
│   │   ├── app_instance.go --- 应用实例管理器
│   │   ├── app_instance_test.go
│   │   ├── host.go --- 主机管理器
│   │   ├── k8s_install.go --- k8s管理器
│   │   └── README.md
│   ├── orm --- 实现实体对象的数据库读写
│   │   ├── conversion.go --- 实体对象跨版本转换方法(未实现)
│   │   ├── core
│   │   │   ├── api_object.go --- 实体对象接口定义
│   │   │   ├── base_object.go --- 通用实体对象定义
│   │   │   ├── common.go --- 实体对象所用到的一些公共常量与方法
│   │   │   ├── meta.go --- 实体对象元数据定义
│   │   │   ├── sort.go --- 实体对象排序方法实现
│   │   │   └── status.go --- 实体对象状态定义
│   │   ├── helper.go --- 整合各版本实体对象存储器的调用
│   │   ├── README.md
│   │   ├── registry.go --- 实体对象存储器接口定义
│   │   ├── runtime --- 各种实体资源的跨版本中转对象定义(未使用)
│   │   └── v1 --- 各种v1版本的实体资源定义
│   │       ├── app.go --- 应用实体对象和存储器定义
│   │       ├── app_instance.go --- 应用实例实体对象和存储器定义
│   │       ├── app_instance_test.go
│   │       ├── app_test.go
│   │       ├── audit.go --- 审计日志实体对象和存储器定义
│   │       ├── audit_test.go
│   │       ├── common.go --- 实体对象中的部分公共资源定义
│   │       ├── configmap.go
│   │       ├── event.go --- 事件日志实体对象和存储器定义
│   │       ├── event_test.go
│   │       ├── gen_k8s_yaml.go --- k8s集群实体对象和存储器定义
│   │       ├── gen_k8s_yaml_test.go
│   │       ├── helper.go --- 整合各个实体对象存储器的调用
│   │       ├── host.go --- 主机实体对象和存储器定义
│   │       ├── host_test.go
│   │       ├── job.go --- 任务实体对象和存储器定义
│   │       ├── job_test.go
│   │       └── registry.go --- v1版本通用实体对象存储器实现
│   ├── schedule
│   │   ├── scheduler.go --- 实现任务调度器
│   │   ├── scheduler_test.go
│   │   └── worker.go --- 实现任务工作器
│   ├── setting
│   │   └── setting.go --- 配置文件加载
│   └── util --- 其他的工具方法
│       ├── jwt.go --- jwt token生成
│       ├── md5.go --- md5计算
│       ├── remote_ssh.go --- 通过ssh向远程主机执行指令
│       ├── remote_ssh_test.go
│       ├── tailf.go --- 日志文件侦听
│       ├── tailf_test.go
│       └── util.go
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
