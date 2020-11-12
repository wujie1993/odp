package registry

import (
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
)

var storageVersion map[core.GK]string
var storageRegistry map[core.GVK]ApiObjectRegistry

func init() {
	storageVersion = make(map[core.GK]string)
	storageRegistry = make(map[core.GVK]ApiObjectRegistry)
}

// RegisterStorageVersion 注册数据库中实际存储的对象版本
func RegisterStorageVersion(gk core.GK, apiVersion string) error {
	registeredVersion, ok := storageVersion[gk]
	if ok {
		return e.Errorf("%+v already register with version %s", gk, registeredVersion)
	}
	storageVersion[gk] = apiVersion
	return nil
}

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

func SetConvertByBytesFunc(f core.ConvertByBytesFunc) {
	convertByBytes = f
}

func SetNewByGVKFunc(f core.NewByGVKFunc) {
	newByGVK = f
}
