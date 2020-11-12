package v1

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
)

type Pkg struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            PkgSpec
}

type PkgSpec struct {
	Desc      string
	Module    string
	Version   string
	Platform  string
	Provision string
	Synced    bool
	Author    string
	Images    []string
}

func (obj Pkg) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

func (obj *Pkg) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

func (obj Pkg) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

type PkgRegistry struct {
	registry.Registry
}

func NewPkg() *Pkg {
	host := new(Pkg)
	host.Init(ApiVersion, core.KindPkg)
	return host
}

func NewPkgRegistry() PkgRegistry {
	return PkgRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindPkg), false),
	}
}
