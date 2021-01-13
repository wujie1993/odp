## dpctl 

用于管理部署平台资源的命令行工具

## 编译

在项目根路径上执行

```
make ctl
```

生成的二进制文件存放于build/dpctl

## 使用示例

具体使用说明请参考`dpctl --help`

### 获取资源

```
# 获取所有主机(可使用host或node), 并以表格形式输出(默认为table格式，可不传--format参数)
dpctl get node --format table

# 获取所有应用, 并以yaml格式输出, 多个资源使用---分隔
dpctl get app --format yaml

# 获取所有配置字典(可使用configmap或cm), 并以json格式输出
dpctl get cm --format json

# 获取所有实例(可使用appinstance或ins), 并以json缩进格式输出
dpctl get ins --format json-pretty
```

> 修改资源时，建议使用`dpctl get [RESOURCE] [NAME] --format yaml > [FILENAME]`指令导出至本地进行编辑

### 创建资源

```
# 从yaml格式文件中创建资源，多个资源使用---做分隔符
dpctl create -f hosts.yml

# 从json格式文件中创建资源，多个资源可使用数组，数组中的资源类型可不同
dpctl create -f configmaps.json
```

> 当资源已存在时会报错

### 应用资源(推荐)

```
# 从yaml格式文件中应用资源，多个资源使用---做分隔符
dpctl apply -f hosts.yml

# 从json格式文件中应用资源，多个资源可使用数组，数组中的资源类型可不同
dpctl apply -f configmaps.json
```

> 当资源不存在时会创建，当资源已存在时会更新

### 删除资源

```
# 删除yaml文件中的资源
dpctl delete -f hosts.yml

# 删除指定名称的资源
dpctl delete ins mysql-e9fsb9sdf9
```

### 主机插件管理

```
# 安装主机插件
dpctl hostplugin install --host host-172.25.21.99 --plugin-name nodejs --plugin-version 10.15.3

# 卸载主机插件
dpctl hostplugin uninstall --host host-172.25.21.99 --plugin-name nodejs
```

> 已安装的主机插件无法再次安装，已卸载的主机插件也无法再次卸载，如果需要强制执行，可以附加参数--force

## 推荐用法

### 主机

主机资源可使用以下yaml文件模板创建

```yaml
---
ApiVersion: v2
Kind: host
Metadata:
  Annotations:
    # 主机的显示别名
    ShortName: 172.25.21.32
  # 主机的唯一标识名
  Name: host-172.25.21.32
Spec:
  # SSH连接配置
  SSH:
    Host: 172.25.21.32
    Password: *****
    Port: 22
    User: root
```

将上方内容保存至hosts.yml文件中，使用`dpctl apply -f hosts.yml`创建主机

### 应用实例

应用实例的管理分几个动作，创建，安装，更新，升级和卸载

#### 创建并安装

可使用如下yaml文件模板创建应用实例，其中包含两个部分，应用实例和配置字典，配置字典存放应用实例中的配置文件内容

```yaml
---
ApiVersion: v2
Kind: appInstance
Metadata:
  Annotations:
    ShortName: my-geoserver
  Name: geoserver-16dae52e
  Namespace: default
Spec:
  # 对应用实例采取的操作，install(安装)，configure(配置)，upgrade(升级)和uninstall(卸载)，留空时则只创建记录而不实际安装
  Action: "install"
  AppRef:
    Name: geoserver
    Version: v1.0.0-10
  Category: customize
  Global:
    Args:
    - Name: nacos_ip
      Value: ""
    - Name: deploy_dir
      Value: /opt/prophet/geoserver
    - Name: logs_prefix
      Value: /data/logs/prophet/geoserver
  LivenessProbe:
    InitialDelaySeconds: 0
    PeriodSeconds: 60
    TimeoutSeconds: 60
  Modules:
  # 模块geoserver配置
  - AppVersion: v1.0.0-10
    Name: geoserver
    Replicas:
    - HostRefs:
      - host-172.25.21.32
      ConfigMapRef:
        # 配置文件所关联的配置字典
        Name: configs-app-geoserver-v1.0.0-10-geoserver-40d7fa7c
        Namespace: default
  # 模块mapserver配置
  - AppVersion: v1.0.0-softLink-521
    Name: mapserver
    Replicas:
    - HostRefs:
      - host-172.25.21.32
---
ApiVersion: v1
Kind: configMap
Metadata:
  Name: configs-app-geoserver-v1.0.0-10-geoserver-40d7fa7c
  Namespace: default
Data:
  # key表示配置文件名，value表示配置文件内容
  start.ini: |-
    #
    # Jetty configuration, taken originally from jetty-9.2.13.v20150730-distribution.zip
    #
    
    # --------------------------------------- 
    # Module: server
    --module=server
    
    # minimum number of threads
    threads.min=10
    # maximum number of threads
    threads.max=200
    # thread idle timeout in milliseconds
    threads.timeout=60000
    # buffer size for output
    jetty.output.buffer.size=32768
    # request header buffer size
    jetty.request.header.size=8192
    # response header buffer size
    jetty.response.header.size=8192
    # should jetty send the server version header?
    jetty.send.server.version=true
    # should jetty send the date header?
    jetty.send.date.header=false
    # What host to listen on (leave commented to listen on all interfaces)
    #jetty.host=myhost.com
    # Dump the state of the Jetty server, components, and webapps after startup
    jetty.dump.start=false
    # Dump the state of the Jetty server, before stop
    jetty.dump.stop=false
    # Enable delayed dispatch optimisation
    jetty.delayDispatchUntilContent=false
    
    # --------------------------------------- 
    # Module: servlets
    --module=servlets
    
    # --------------------------------------- 
    # Module: deploy
    --module=deploy
    
    # Monitored Directory name (relative to jetty.base)
    # jetty.deploy.monitoredDirName=webapps
    
    # --------------------------------------- 
    # Module: websocket
    #--module=websocket
    
    # --------------------------------------- 
    # Module: ext
    #--module=ext
    
    # --------------------------------------- 
    # Module: resources
    --module=resources
    
    # --------------------------------------- 
    # Module: http
    --module=http
    
    # HTTP port to listen on
    jetty.port=30880
    
    # HTTP idle timeout in milliseconds
    http.timeout=30000
    
    # HTTP Socket.soLingerTime in seconds. (-1 to disable)
    # http.soLingerTime=-1
    
    # Parameters to control the number and priority of acceptors and selectors
    # http.selectors=1
    # http.acceptors=1
    # http.selectorPriorityDelta=0
    # http.acceptorPriorityDelta=0
    
    # --------------------------------------- 
    # Module: webapp
    --module=webapp

```

将以上内容编辑并保存至geoserver.yml文件中

使用命令`dpctl apply -f geoserver.yml`创建并安装实例

#### 配置更新

将geoserver.yml中应用实例的`.Spec.Action`值改为`configure`, 修改配置字典中的配置文件内容

使用命令`dpctl apply -f geoserver.yml`更新应用实例

#### 卸载

将geoserver.yml中应用实例的`.Spec.Action`值改为`uninstall`

使用命令`dpctl apply -f geoserver.yml`卸载应用实例
