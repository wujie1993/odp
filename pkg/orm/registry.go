package orm

import (
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/orm/v2"
)

// 注册实体对象存储器, 需要切换存储器版本时在此处修改
func Init() {
	// 注册实体对象的存储版本
	registry.RegisterStorageVersion(core.GK{Group: core.Group, Kind: core.KindAppInstance}, v2.ApiVersion)
	registry.RegisterStorageVersion(core.GK{Group: core.Group, Kind: core.KindEvent}, v1.ApiVersion)
	registry.RegisterStorageVersion(core.GK{Group: core.Group, Kind: core.KindJob}, v2.ApiVersion)

	// 注册实体对象存储器，用于数据迁移
	registry.RegisterStorageRegistry(v1.NewAppRegistry())
	registry.RegisterStorageRegistry(v1.NewAppInstanceRegistry())
	registry.RegisterStorageRegistry(v1.NewAuditRegistry())
	registry.RegisterStorageRegistry(v1.NewConfigMapRegistry())
	registry.RegisterStorageRegistry(v1.NewEventRegistry())
	registry.RegisterStorageRegistry(v1.NewK8sConfigRegistry())
	registry.RegisterStorageRegistry(v1.NewGPURegistry())
	registry.RegisterStorageRegistry(v1.NewHostRegistry())
	registry.RegisterStorageRegistry(v1.NewJobRegistry())
	registry.RegisterStorageRegistry(v1.NewPkgRegistry())
	registry.RegisterStorageRegistry(v2.NewAppInstanceRegistry())
	registry.RegisterStorageRegistry(v2.NewJobRegistry())

	migrate()
}

// migrate 用于升级迁移数据
func migrate() {
	// 将归属于命名空间下的资源存储路径进行迁移，从/registry/namespaces/<namespace>/<kind>/<name>迁移至/registry/<kind>/<namespace>/<name>，对v1.2.0及以下版本生效
	registry.MigrateNamespacedObjects()

	// 将数据库中的所有对象更新成Schema中所注册的存储版本
	registry.MigrateStorageVersion()
}
