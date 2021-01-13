package v2

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
)

func (obj Host) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

func (obj *Host) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

func (obj Host) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

type HostRegistry struct {
	registry.Registry
}

func NewHost() *Host {
	host := new(Host)
	host.Init(ApiVersion, core.KindHost)
	return host
}

func NewHostRegistry() HostRegistry {
	r := HostRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindHost), false),
	}
	r.SetDefaultFinalizers([]string{
		core.FinalizerCleanRefGPU,
		core.FinalizerCleanRefEvent,
	})
	return r
}
