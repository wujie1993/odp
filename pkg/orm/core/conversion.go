package core

type NewByGVKFunc func(gvk GVK) (ApiObject, error)
type ConvertFunc func(srcObj ApiObject, dstGVK GVK) (ApiObject, error)
type ConvertByBytesFunc func(srcObjBytes []byte, dstGVK GVK) (ApiObject, error)

// GVK Group Version Kind简称
type GVK struct {
	Group      string
	ApiVersion string
	Kind       string
}

// GK Group Kind简称
type GK struct {
	Group string
	Kind  string
}

// VK Version Kind简称
type VK struct {
	ApiVersion string
	Kind       string
}

// Conversion 结构转换方法注册器
type Conversion struct {
	convertFuncs map[gvkPair]ConvertFunc
}

type gvkPair struct {
	srcGVK GVK
	dstGVK GVK
}

func (c *Conversion) SetConversionFunc(srcGVK GVK, dstGVK GVK, convertFunc ConvertFunc) {
	c.convertFuncs[gvkPair{srcGVK: srcGVK, dstGVK: dstGVK}] = convertFunc
}

func (c *Conversion) GetConversionFunc(srcGVK GVK, dstGVK GVK) (ConvertFunc, bool) {
	convertFunc, ok := c.convertFuncs[gvkPair{srcGVK: srcGVK, dstGVK: dstGVK}]
	return convertFunc, ok
}

func NewConversion() Conversion {
	return Conversion{
		convertFuncs: make(map[gvkPair]ConvertFunc),
	}
}
