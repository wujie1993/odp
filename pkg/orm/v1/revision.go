package v1

import (
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
)

type Revision struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	ResourceRef     ResourceRef
	Revision        int
	Data            string
}

func (obj Revision) SpecEncode() ([]byte, error) {
	return nil, nil
}

func (obj *Revision) SpecDecode(data []byte) error {
	return nil
}

func (obj Revision) SpecHash() string {
	return ""
}

type RevisionRegistry struct {
	registry.Registry
}

func NewRevision() *Revision {
	revision := new(Revision)
	revision.Init(ApiVersion, core.KindRevision)
	return revision
}

func NewRevisionRegistry() RevisionRegistry {
	r := RevisionRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindRevision), false),
	}
	return r
}
