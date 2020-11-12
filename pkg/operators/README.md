# 模块管理器

管理器是一个独立模块的自治系统，通常会侦听某一种资源的变更，并根据变更的内容执行相应的动作

管理器具有以下特点:

- 单例运行
- 常驻于内存区
- 通过watch方式异步执行动作，与接口解耦
- 分阶段处理，管理器中的handle方法一次只处理资源的一个状态阶段

管理器的工作流程如下:

```mermaid
sequenceDiagram
    operator.Run()->>ETCD数据库: 侦听资源变更为"状态1"
    operator.Run()-->>operator.handle(): 交由handle()处理处于"状态1"的资源
    operator.handle()-->>ETCD数据库: 更新资源为"状态2"
    operator.Run()->>ETCD数据库: 侦听资源变更为"状态2"
    operator.Run()-->>operator.handle(): 交由handle()处理处于"状态2"的资源
    operator.handle()-->>ETCD数据库: 更新资源为"状态n"
    operator.Run()->>ETCD数据库: 侦听资源变更为"状态n"
    operator.Run()-->>operator.handle(): 交由handle()处理处于"状态n"的资源
```

## 主机管理器

主机管理器主要完成以下工作：

1. 检测主机连接状态
2. 采集主机系统信息
3. 下发命令到主机上

## 应用实例管理器

应用实例管理器主要完成以下工作：

1. 安装应用实例
2. 配置应用实例
3. 卸载应用实例

## K8S管理器

K8S管理器主要完成以下工作：

1. K8S集群安装
2. K8S集群卸载
3. K8S节点打标签

## 添加自定义管理器

以应用实例管理器为例

1. 定义管理器

```
package operators

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/ansible"
	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/setting"
)

// 定义管理器
type AppInstanceOperator struct {
        BaseOperator
}

// handle 方法根据应用实例的状态(.status.phase)做相应的处理，并更新状态，原则上handle方法的一次调用只处理一种状态，状态一旦更新则立即退出，交由下一次的handle处理
func (c *AppInstanceOperator) handleAppInstance(ctx context.Context, obj core.ApiObject) {
	appInstance := obj.(*v1.V1AppInstance)
	log.Infof("%s '%s' is %s", appInstance.Kind, appInstance.GetKey(), appInstance.Status.Phase)

	switch appInstance.Status.Phase {
	case core.PhaseWaiting:
                /// TODO: 下方添加处于等待中状态时的处理逻辑 
                /// ...
                /// 逻辑处理结束

		// 更新应用实例状态
		appInstance.Status.SetCondition(core.ConditionTypeInitialized, core.ConditionStatusTrue)
		if _, err := c.helper.V1.AppInstance.Update(appInstance, true); err != nil {
			log.Error(err)
			return
		}

		// 记录事件开始
		if err := o.recordEvent(Event{
			BaseApiObj: appInstance.BaseApiObj,
			Action:     action,
			Msg:        "",
			JobRef:     job.Metadata.Name,
			Phase:      core.PhaseWaiting,
		}); err != nil {
			log.Error(err)
		}
	case core.PhaseUninstalling:
                /// TODO: 下方添加处于卸载中状态时的处理逻辑 
                /// ...
                /// 逻辑处理结束
	case core.PhaseInstalling:
                /// TODO: 下方添加处于安装中状态时的处理逻辑 
                /// ...
                /// 逻辑处理结束
	case core.PhaseInstalled, core.PhaseUninstalled, core.PhaseFailed:
                /// TODO: 下方添加处于已安装，已卸载或失败状态时的处理逻辑 
                /// ...
                /// 逻辑处理结束
	default:
		// 处于其他状态，如果内容体发生更新则将状态置为等待中
		if hash, ok := c.hashMap[appInstance.GetKey()]; ok && hash != appInstance.SpecHash() {
			if _, err := c.helper.V1.AppInstance.UpdateStatusPhase(appInstance.Metadata.Namespace, appInstance.Metadata.Name, core.PhaseWaiting); err != nil {
				log.Error(err)
			}
			return
		}
	}
}

// NewAppInstanceOperator 实例化应用实例管理器
func NewAppInstanceOperator(namespace string) *AppInstanceOperator {
	o := &AppInstanceOperator{
		BaseOperator: BaseOperator{
			namespace: namespace,
			registry:  v1.NewV1AppInstanceRegistry(),
			helper:    orm.GetHelper(),
			objQueue:  make(chan core.ApiObject, 1000),
			hashMap:   make(map[string]string),
		},
	}
	o.BaseOperator.handle = o.handleAppInstance
	o.healthCheckMap = make(map[string]context.CancelFunc)
	return o
}
```

2. 在`main.go`中启动管理器

```
...
func loadPlugins(ctx context.Context) {
        ...

	appInstanceOperator := operators.NewAppInstanceOperator("default")
	go appInstanceOperator.Run(ctx)
}
...
```
