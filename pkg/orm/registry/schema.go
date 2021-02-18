package registry

import (
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
)

var (
	storageVersion  = make(map[core.GK]string)
	storageRegistry = make(map[core.GVK]ApiObjectRegistry)
)

// RegisterStorageVersion 注册数据库中实际存储的对象版本
func RegisterStorageVersion(gk core.GK, apiVersion string) error {
	registeredVersion, ok := storageVersion[gk]
	if ok {
		return e.Errorf("%+v already register with version %s", gk, registeredVersion)
	}
	storageVersion[gk] = apiVersion
	return nil
}

// RegisterStorageRegistry 注册可用于数据迁移的存储器
func RegisterStorageRegistry(registry ApiObjectRegistry) error {
	gvk := registry.GVK()
	registeredRegistry, ok := storageRegistry[gvk]
	if ok {
		return e.Errorf("%+v already register with registry %+v", gvk, reflect.TypeOf(registeredRegistry))
	}
	log.Debugf("register %+v with registry %+v", gvk, reflect.TypeOf(registry))
	storageRegistry[gvk] = registry
	return nil
}

// MigrateNamespacedObjects 将归属于命名空间下的资源存储路径进行迁移，从/registry/namespaces/<namespace>/<kind>/<name>迁移至/registry/<kind>/<namespace>/<name>
func MigrateNamespacedObjects() {
	keyPrefix := core.RegistryPrefix + "/namespaces/"
	kvList, err := db.KV.List(keyPrefix, true)
	if err != nil {
		log.Error(err)
		return
	}

	for key, value := range kvList {
		// 拆分<namespace>/<kind>/<name>
		keyParts := strings.Split(strings.TrimPrefix(key, keyPrefix), "/")
		if len(keyParts) < 3 {
			continue
		}

		// 重组新键名
		newKey := core.RegistryPrefix + "/" + keyParts[1] + "/" + keyParts[0] + "/" + keyParts[2]
		log.Debugf("migrate storage path from %s to %s", key, newKey)

		// 写入新数据
		if err := db.KV.Set(newKey, value); err != nil {
			log.Error(err)
			return
		}

		// 删除旧数据
		if _, err := db.KV.Delete(key); err != nil {
			log.Error(err)
			return
		}
	}
}

// UpgradeStorageVersion 将数据库中的对象转换为注册版本
func MigrateStorageVersion() {
	for gk, apiVersion := range storageVersion {
		gvk := core.GVK{Group: gk.Group, ApiVersion: apiVersion, Kind: gk.Kind}
		registry, ok := storageRegistry[gvk]
		if !ok {
			log.Warnf("%+v hasn't register with any registry", gvk)
			continue
		} else if registry == nil {
			log.Warnf("registry of %+v not found", gvk)
			continue
		}
		if err := registry.MigrateObjects(); err != nil {
			log.Error(err)
		}
	}
}

var convertByBytes core.ConvertByBytesFunc
var newByGVK core.NewByGVKFunc

// SetConverByBytesFunc 设置将字节转换为资源结构的方法实现，该方法在存储器中使用，将未知的序列化数据反序列化为明确类型的资源对象，并转换为指定版本结构的资源对象
func SetConvertByBytesFunc(f core.ConvertByBytesFunc) {
	convertByBytes = f
}

// SetNewByGVKFunc 设置根据GVK生成明确类型资源对象的方法，该方法在数据迁移时使用，根据Group,ApiVersion和Kind信息实例化对应的资源对象
func SetNewByGVKFunc(f core.NewByGVKFunc) {
	newByGVK = f
}
