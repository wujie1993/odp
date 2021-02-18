package orm

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/orm/v2"
)

// InitStorage 初始化底层存储
func InitStorage() {
	// 注册实体对象的存储版本，在后续的迁移数据中会将相同Kind的不同结构版本都转换为同一个版本
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

	// 迁移数据结构
	migrate()

	// 创建初始数据
	initData()
}

// migrate 用于升级迁移数据
func migrate() {
	// 将归属于命名空间下的资源存储路径进行迁移，从/registry/namespaces/<namespace>/<kind>/<name>迁移至/registry/<kind>/<namespace>/<name>，只对v1.2.0及以下版本生效
	registry.MigrateNamespacedObjects()

	// 将数据库中的所有对象更新成Schema中所注册的存储版本
	registry.MigrateStorageVersion()
}

// initData 创建初始数据
func initData() {
	// 创建默认项目空间
	projectRegistry := v1.NewProjectRegistry()
	projectObj, err := projectRegistry.Get(context.TODO(), "", core.DefaultNamespace)
	if err != nil {
		log.Error(err)
		return
	} else if projectObj != nil {
		return
	}
	project := v1.NewProject()
	project.Metadata.Name = "default"
	project.Metadata.Annotations["ShortName"] = "默认项目"
	if _, err := projectRegistry.Create(context.TODO(), project); err != nil {
		log.Error(err)
		return
	}
}
