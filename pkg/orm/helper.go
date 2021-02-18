package orm

import (
	"errors"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/runtime"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/orm/v2"
)

func init() {
	helper = Helper{
		V1: v1.GetHelper(),
		V2: v2.GetHelper(),
	}
}

var helper Helper

// Helper 做为所有资源存储器的统一访问入口
type Helper struct {
	V1 v1.Helper
	V2 v2.Helper
}

func GetHelper() *Helper {
	return &helper
}

// NewByMetaType 根据MetaType实例化资源
func NewByMetaType(metaType core.MetaType) (core.ApiObject, error) {
	return New(core.GVK{
		ApiVersion: metaType.ApiVersion,
		Group:      core.Group,
		Kind:       metaType.Kind,
	})
}

// New 根据GVK实例化资源
func New(gvk core.GVK) (core.ApiObject, error) {
	switch gvk.ApiVersion {
	case v1.ApiVersion:
		return v1.New(gvk.Kind)
	case v2.ApiVersion:
		return v2.New(gvk.Kind)
	case "":
		return runtime.New(gvk.Kind)
	default:
		return nil, errors.New("unknown apiVersion")
	}
}
