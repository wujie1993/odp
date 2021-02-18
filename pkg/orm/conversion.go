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

// Conversion 资源对象结构转换方法注册器
type Conversion struct {
	// 存储资源的版本结构与运行时结构的互相转换方法
	versionedConvertFunc map[string]core.ConvertFunc
}

func init() {
	conversion = Conversion{
		versionedConvertFunc: make(map[string]core.ConvertFunc),
	}
	// 注册v1版本结构与runtime结构的互相转换方法
	conversion.versionedConvertFunc[v1.ApiVersion] = v1.Convert
	// 注册v2版本结构与runtime结构的互相转换方法
	conversion.versionedConvertFunc[v2.ApiVersion] = v2.Convert

	// 注入转换方法所需使用的方法，使用方法注入而非直接在目标包中定义方法的原因是为了避免出现包的循环依赖
	registry.SetNewByGVKFunc(New)
	registry.SetConvertByBytesFunc(ConvertByBytes)
}

// Convert 转换相同Group相同Kind的不同版本对象结构,转换方式分以下两种:
// 1. 直接转换: versioned -> unversioned || unversioned -> versioned
// 2. 间接转换: versioned -> unversioned -> versioned
// 转换对象结构时所使用的转换方法在Conversion中获取，当方法不存在时会转换失败
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
// 例如，从数据库中读取的是v1结构的序列化数据，而目标的结构是v2。
// 首先会对读取出的数据进行反序列化为v1结构的对象，再通过Convert方法将v1结构的对象转换为v2结构的对象。
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
