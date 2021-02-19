## ORM 使用说明

### 说明

数据库中的存储对象首先按版本号划分，如：v1，v2等。再按照类型划分，如：host, job等。

**多版本互相转换**

同一种资源类型的多个版本的数据是可以共存并互相转换的，为了实现这一特性，需要额外定义一个runtime版本结构，runtime版本结构作为其他版本间互相转换的中转结构，即当v1版本结构要转换为v2版本时，需要经过v1->runtime->v2，当v2版本结构奥转换为v1版本时，需要经过v2->runtime->v1。

为何要经过runtime中转而不直接v1->v2和v2->v1？

可以从需要实现的结构转换方法数量考虑。当host有n个版本的结构，要实现各个版本间的结构转换，需要实现n(n-1)个转换方法；而使用runtime间接转换的方式，需要实现2n个。可通过计算得出当n>3时，直接转换的方式需要实现的转换方法数量会超过runtime间接转换的方式。

实现跨版本转换的方法如下：

假设实现job的v1和v2版本互相转换，在./v1/conversion.go中实现v1->runtime和runtime->v1的转换方法，在./v2/conversion.go中实现v2->runtime和runtime->v2的转换方法。

```
func init(){
        ...
	registerConversionFunc(core.VK{
		ApiVersion: ApiVersion,
		Kind:       core.KindJob,
	}, core.VK{
		Kind: core.KindJob,
	}, convertCoreV1JobToCoreRuntimeJob)

	registerConversionFunc(core.VK{
		Kind: core.KindJob,
	}, core.VK{
		ApiVersion: ApiVersion,
		Kind:       core.KindJob,
	}, convertCoreRuntimeJobToCoreV1Job)
        ...
}

// v1->runtime
func convertCoreV1JobToCoreRuntimeJob(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
        ...
}

// runtime->v1
func convertCoreRuntimeJobToCoreV1Job(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
        ...
}
```

**存储注册**

默认情况下同一种资源在数据库中可以存储多个版本，在读取时根据所使用的存储器版本进行结构转换，而结构转换是有性能损耗的。

打个比方，host资源可以在数据库中同时存储v1和v2版本结构的数据，当使用v2版本的host存储器读取host数据时，会同时获取到v1和v2版本的数据，其中v1版本的数据会转换为v2版本，而v2版本的数据无需转换。

(待实现)为了减小读取性能的损耗，可以通过存储注册的方式将资源的一个版本注册为存储版本，使用任何版本的存储器写入数据时，都会将结构转换为注册的版本

注册存储版本的方法如下：

编辑./registry.go

```golang
func Init() {
        ...
        // 注册为Job资源使用v2版本结构作为存储版本
	registry.RegisterStorageVersion(core.GK{Group: core.Group, Kind: core.KindJob}, v2.ApiVersion)	

        // 注册结构转换所需使用的各个版本存储器，在数据迁移时会使用到
        registry.RegisterStorageRegistry(v1.NewJobRegistry())
        registry.RegisterStorageRegistry(v2.NewJobRegistry())
        ...
}
```

> 服务端在启动时会执行数据迁移，将同一类型的结构数据统一转换为注册的存储版本

**代码生成**

存储对象和存储器中的部分方法可通过代码生成器生成

对于继承了"github.com/wujie1993/waves/pkg/orm/core".BaseApiObj的资源对象实现了以下方法的自动生成

- DeepCopy
- DeepCopyInto
- FromJSON
- ToJSON
- FromYAML
- ToYAML
- Sha256

对于继承了"github.com/wujie1993/waves/pkg/orm/registry".Registry的存储器实现了helper的自动封装

代码生成命令

```
make gen
```

> 自动生成的代码文件名以zz_generated开头

### 资源对象的定义

以创建`Host`对象资源为例：

1. 在`v1/types.go`文件中定义对象的字段结构

```golang
package v1

import (
	"crypto/sha256"
	"encoding/json"
        "fmt"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
)

...

// Host orm对象资源，请将自定义结构字段补充于.Spec中
type Host struct {
        // 每个资源对象都继承于基础对象
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec HostSpec
}

type HostSpec struct {
	SSH HostSSH
}

type HostSSH struct {
	Host   string
	User   string
	Passwd string
}

// SpecHash 计算当前资源对象的.spec内容哈希值，作为对象是否发生更新的判断依据
func (obj Host) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// SpecEncode 序列化Spec字段的内容
func (obj Host) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec.SSH)
}

// SpecDecode 反序列化Spec字段的内容
func (obj *Host) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec.SSH)
}

// NewHost 用于实例化一个新的资源对象
func NewHost() *Host {
	host := new(Host)
	host.Init(ApiVersion, core.KindHost)
	return host
}

...
```

2. 在`v1/registries.go`中定义资源对象存储器，用于为特定的资源对象做数据库读写


```golang
...

// HostRegistry Host资源所对应的对象存储器 
type HostRegistry struct {
        // 每个资源对象存储器都继承于通用存储器，其中已经实现了常用的CRUD方法
	registry.Registry
}


// NewHostRegistry 用于实例化一个新的资源对象数据库操作器
func NewHostRegistry() *HostRegistry {
	r := &HostRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindHost), false),
	}
        return r
}
```

3. 在`core/common.go`中添加该结构对应的常量

```golang
...
const KindHost = "host"
...
```

4. 执行代码生成命令

```bash
make gen
```

### 使用资源对象存储器进行数据读写

1. 初始化对象操作器

```
helper := orm.GetHelper()
```

2. 对象的CRUD操作

```golang
// 创建资源
helper.V1.Host.Create(context.TODO(), host) 

// 获取资源
helper.V1.Host.Get(context.TODO(), namespace, name)

// 更新资源
helper.V1.Host.Update(context.TODO(), host)

// 列举资源
helper.V1.Host.List(context.TODO(), namespace)

// 删除资源
helper.V1.Host.Delete(context.TODO(), namespace, name)

// 侦听资源变动
helper.V1.Host.Watch(ctx, namespace, name)
helper.V1.Host.GetWatch(ctx, namespace, name)
helper.V1.Host.ListWatch(ctx, namespace)
```

> 以上方法返回的资源对象均为core.ApiObject接口，需要获取其中的内容需要做类型推断,如：host:=obj.(*v1.Host)

### 通过钩子方法在数据读写过程中注入自定义逻辑

在资源对象读写的过程中，我们往往需要加入一些自定义的逻辑，例如：字段校验，字段填充和外部探针等。在通用资源对象存储器中实现了钩子方法的注入，可以在数据读写的各个阶段中注入自定义的逻辑。目前支持以下的钩子方法：

- ValidateHook 字段校验钩子，在Create和Update前执行
- MutateHook 字段填充钩子，在Create和Update前执行
- PreCreateHook 创建前置钩子，在Create前执行
- PreUpdateHook 更新前置钩子，在Update前执行
- PreDeleteHook 删除前置钩子，在Delete前执行
- PostCreateHook 创建后置钩子，在Create后执行
- PostUpdateHook 更新后置钩子，在Update后执行
- PostDeleteHook 删除后置钩子，在Delete后执行

所有的钩子方法都使用`type HookFunc func(obj core.ApiObject) error`结构定义，通过SetXXXHook方法注入到存储器中，其中传入参数`obj core.ApiObject`为要发生数据读写的资源对象，通过类型推断后（如：`host := obj.(*v1.Host)`）方可使用，返回参数为`error`，当需要中断整个读写过程时，需要返回非nil值。示例如下：

在`v1/registries.go`中，为Host资源添加自定义字段校验逻辑

```
...

func hostValidate(obj core.ApiObject) error {
        host := obj.(*v1.Host)
        if len(host.Spec.SSH.Password) <= 6 {
                return e.Errorf("密码长度应该超过 7 位")
        }
        return nil
}

func NewHostRegistry() *HostRegistry {
	r := &HostRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindHost), false),
	}
        // 注册字段校验钩子
        r.SetValidateHook(hostValidate)
        return r
}

...
```
