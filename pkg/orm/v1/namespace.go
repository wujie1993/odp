package v1

import (
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
)

type Namespace struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
}

func (obj Namespace) SpecEncode() ([]byte, error) {
	return nil, nil
}

func (obj *Namespace) SpecDecode(data []byte) error {
	return nil
}

func (obj Namespace) SpecHash() string {
	return ""
}

type NamespaceRegistry struct {
	registry.Registry
}

func NewNamespace() *Namespace {
	ns := new(Namespace)
	ns.Init(ApiVersion, core.KindNamespace)
	return ns
}

func NewNamespaceRegistry() NamespaceRegistry {
	r := NamespaceRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindNamespace), false),
	}
	r.SetDefaultFinalizers([]string{
		core.FinalizerCleanRefConfigMap,
	})
	return r
}
