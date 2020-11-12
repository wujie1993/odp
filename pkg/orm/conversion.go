package orm

import (
	"encoding/json"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/orm/v2"
)

var conversion Conversion

type Conversion struct {
	versionedConvertFunc map[string]core.ConvertFunc
}

func init() {
	conversion = NewConversion()
	conversion.versionedConvertFunc[v1.ApiVersion] = v1.Convert
	conversion.versionedConvertFunc[v2.ApiVersion] = v2.Convert

	// 依赖注入
	registry.SetNewByGVKFunc(New)
	registry.SetConvertByBytesFunc(ConvertByBytes)
}

func NewConversion() Conversion {
	return Conversion{
		versionedConvertFunc: make(map[string]core.ConvertFunc),
	}
}

// Convert 转换相同Group相同Kind的不同版本对象结构,转换方式分以下两种:
// 1. 直接转换: versioned -> unversioned || unversioned -> versioned
// 2. 间接转换: versioned -> unversioned -> versioned
func Convert(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
	srcGVK := srcObj.GetGVK()

	if srcGVK.Group != dstGVK.Group {
		err := e.Errorf("Convert %s to %+v failed. Group not match", srcObj.GetKey(), dstGVK)
		log.Error(err)
		return nil, err
	}
	if srcGVK.Kind != dstGVK.Kind {
		err := e.Errorf("Convert %s to %+v failed. Kind not match", srcObj.GetKey(), dstGVK)
		log.Error(err)
		return nil, err
	}

	if srcGVK == dstGVK {
		// 源与目标结构一致，直接返回源目标对象
		return srcObj, nil
	}

	if srcGVK.ApiVersion != "" && dstGVK.ApiVersion != "" {
		// versioned -> unversioned -> versioned
		// 转为运行时结构
		convertFunc, ok := conversion.versionedConvertFunc[srcGVK.ApiVersion]
		if !ok {
			err := e.Errorf("Convert %s to %+v failed. Source object can't convert to runtime object", srcObj.GetKey(), dstGVK)
			log.Error(err)
			return nil, err
		}
		rtObj, err := convertFunc(srcObj, core.GVK{Group: core.Group, Kind: dstGVK.Kind})
		if err != nil {
			log.Error(err)
			return nil, err
		}

		if rtObj.GetGVK().ApiVersion != "" {
			err := e.Errorf("%+v %+v not runtime object", reflect.TypeOf(rtObj), rtObj)
			log.Error(err)
			return nil, err
		}

		// 转为目标结构
		convertFunc, ok = conversion.versionedConvertFunc[dstGVK.ApiVersion]
		if !ok {
			err := e.Errorf("Convert %s to %+v failed. Runtime object can't convert to destnation object", srcObj.GetKey(), dstGVK)
			log.Error(err)
			return nil, err
		}
		dstObj, err := convertFunc(rtObj, dstGVK)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		return dstObj, nil
	} else if srcGVK.ApiVersion != "" && dstGVK.ApiVersion == "" {
		// versioned -> unversioned
		convertFunc, ok := conversion.versionedConvertFunc[srcGVK.ApiVersion]
		if !ok {
			err := e.Errorf("Convert %s to %+v failed.", srcObj.GetKey(), dstGVK)
			log.Error(err)
			return nil, err
		}
		dstObj, err := convertFunc(srcObj, dstGVK)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		return dstObj, nil
	} else if srcGVK.ApiVersion == "" && dstGVK.ApiVersion != "" {
		// unversioned -> versioned
		convertFunc, ok := conversion.versionedConvertFunc[dstGVK.ApiVersion]
		if !ok {
			err := e.Errorf("Convert %s to %+v failed.", srcObj.GetKey(), dstGVK)
			log.Error(err)
			return nil, err
		}
		dstObj, err := convertFunc(srcObj, dstGVK)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		return dstObj, nil
	}

	return nil, e.Errorf("Convert %s to %+v failed. Unsupported conversion", srcObj.GetKey(), dstGVK)
}

// ConvertByBytes 转换相同Group相同Kind的不同版本对象结构
func ConvertByBytes(srcObjBytes []byte, dstGVK core.GVK) (core.ApiObject, error) {
	metaType := new(core.MetaType)
	if err := json.Unmarshal(srcObjBytes, metaType); err != nil {
		log.Error(err)
		return nil, err
	}

	srcObj, err := New(core.GVK{Group: core.Group, ApiVersion: metaType.ApiVersion, Kind: metaType.Kind})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if err := json.Unmarshal(srcObjBytes, srcObj); err != nil {
		log.Error(err)
		return nil, err
	}
	return Convert(srcObj, dstGVK)
}
