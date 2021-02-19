package v2

import (
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/runtime"
)

var conversion core.Conversion

func init() {
	// 实例化并注册v2结构与运行时结构的互相转换方法
	conversion = core.NewConversion()

	registerConversionFunc(core.VK{
		Kind: core.KindAppInstance,
	}, core.VK{
		ApiVersion: ApiVersion,
		Kind:       core.KindAppInstance,
	}, convertCoreRuntimeAppInstanceToCoreV2AppInstance)

	registerConversionFunc(core.VK{
		ApiVersion: ApiVersion,
		Kind:       core.KindAppInstance,
	}, core.VK{
		Kind: core.KindAppInstance,
	}, convertCoreV2AppInstanceToCoreRuntimeAppInstance)

	registerConversionFunc(core.VK{
		Kind: core.KindJob,
	}, core.VK{
		ApiVersion: ApiVersion,
		Kind:       core.KindJob,
	}, convertCoreRuntimeJobToCoreV2Job)

	registerConversionFunc(core.VK{
		ApiVersion: ApiVersion,
		Kind:       core.KindJob,
	}, core.VK{
		Kind: core.KindJob,
	}, convertCoreV2JobToCoreRuntimeJob)

	registerConversionFunc(core.VK{
		Kind: core.KindHost,
	}, core.VK{
		ApiVersion: ApiVersion,
		Kind:       core.KindHost,
	}, convertCoreRuntimeHostToCoreV2Host)

	registerConversionFunc(core.VK{
		ApiVersion: ApiVersion,
		Kind:       core.KindHost,
	}, core.VK{
		Kind: core.KindHost,
	}, convertCoreV2HostToCoreRuntimeHost)
}

func newGVK(kind string) core.GVK {
	return core.GVK{
		Group:      core.Group,
		ApiVersion: ApiVersion,
		Kind:       kind,
	}
}

// registerConversionFunc 注册结构转换方法
func registerConversionFunc(srcVK core.VK, dstVK core.VK, convertFunc core.ConvertFunc) {
	conversion.SetConversionFunc(core.GVK{
		Group:      core.Group,
		ApiVersion: srcVK.ApiVersion,
		Kind:       srcVK.Kind,
	}, core.GVK{
		Group:      core.Group,
		ApiVersion: dstVK.ApiVersion,
		Kind:       dstVK.Kind,
	}, convertFunc)
}

// Convert 将v2版本结构与运行时结构互相转换
func Convert(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
	srcGVK := srcObj.GetGVK()

	if srcGVK == dstGVK {
		// 源与目标结构一致，直接返回源目标对象
		return srcObj, nil
	}

	if (srcGVK.ApiVersion == "" && dstGVK.ApiVersion == "") || (srcGVK.ApiVersion != "" && dstGVK.ApiVersion != "") {
		return nil, e.Errorf("Convert %s to %+v failed. Unsupported conversion", srcObj.GetKey(), dstGVK)
	}

	log.Tracef("convert %+v %+v from %+v to %+v", reflect.TypeOf(srcObj), srcObj, srcGVK, dstGVK)
	// 直接转换
	convertFunc, ok := conversion.GetConversionFunc(srcGVK, dstGVK)
	if !ok {
		return nil, e.Errorf("Convert %s to %+v failed. Convert function not found", srcObj.GetKey(), dstGVK)
	}
	return convertFunc(srcObj, dstGVK)
}

// convertCoreRuntimeAppInstanceToCoreV2AppInstance 将应用实例运行时结构转换为v2结构
func convertCoreRuntimeAppInstanceToCoreV2AppInstance(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
	_, ok := srcObj.(*runtime.AppInstance)
	if !ok {
		return nil, e.Errorf("mismatch with type of source object")
	}
	dstObj, err := New(dstGVK.Kind)
	if err != nil {
		return nil, err
	}
	if err := core.DeepCopy(srcObj, dstObj); err != nil {
		return nil, err
	}
	dstObj.SetGVK(dstGVK)
	return dstObj, nil
}

// convertCoreV2AppInstanceToCoreRuntimeAppInstance 将应用实例v2结构转换为运行时结构
func convertCoreV2AppInstanceToCoreRuntimeAppInstance(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
	_, ok := srcObj.(*AppInstance)
	if !ok {
		return nil, e.Errorf("mismatch with type of source object")
	}
	dstObj, err := runtime.New(dstGVK.Kind)
	if err != nil {
		return nil, err
	}
	if err := core.DeepCopy(srcObj, dstObj); err != nil {
		return nil, err
	}
	dstObj.SetGVK(dstGVK)
	return dstObj, nil
}

// convertCoreRuntimeJobToCoreV2Job 将任务运行时结构转换为v2结构
func convertCoreRuntimeJobToCoreV2Job(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
	_, ok := srcObj.(*runtime.Job)
	if !ok {
		return nil, e.Errorf("mismatch with type of source object")
	}
	dstObj, err := New(dstGVK.Kind)
	if err != nil {
		return nil, err
	}
	if err := core.DeepCopy(srcObj, dstObj); err != nil {
		return nil, err
	}
	dstObj.SetGVK(dstGVK)
	return dstObj, nil
}

// convertCoreV2JobToCoreRuntimeJob 将任务v2结构转换为运行时结构
func convertCoreV2JobToCoreRuntimeJob(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
	_, ok := srcObj.(*Job)
	if !ok {
		return nil, e.Errorf("mismatch with type of source object")
	}
	dstObj, err := runtime.New(dstGVK.Kind)
	if err != nil {
		return nil, err
	}
	if err := core.DeepCopy(srcObj, dstObj); err != nil {
		return nil, err
	}
	dstObj.SetGVK(dstGVK)
	return dstObj, nil
}

// convertCoreRuntimeHostToCoreV2Host 将主机运行时结构转换为v2结构
func convertCoreRuntimeHostToCoreV2Host(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
	_, ok := srcObj.(*runtime.Host)
	if !ok {
		return nil, e.Errorf("mismatch with type of source object")
	}
	dstObj, err := New(dstGVK.Kind)
	if err != nil {
		return nil, err
	}
	if err := core.DeepCopy(srcObj, dstObj); err != nil {
		return nil, err
	}
	dstObj.SetGVK(dstGVK)
	return dstObj, nil
}

// convertCoreV2HostToCoreRuntimeHost 将主机v2结构转换为运行时结构
func convertCoreV2HostToCoreRuntimeHost(srcObj core.ApiObject, dstGVK core.GVK) (core.ApiObject, error) {
	_, ok := srcObj.(*Host)
	if !ok {
		return nil, e.Errorf("mismatch with type of source object")
	}
	dstObj, err := runtime.New(dstGVK.Kind)
	if err != nil {
		return nil, err
	}
	if err := core.DeepCopy(srcObj, dstObj); err != nil {
		return nil, err
	}
	dstObj.SetGVK(dstGVK)
	return dstObj, nil
}
