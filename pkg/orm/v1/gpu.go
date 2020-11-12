package v1

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
)

type GPU struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            GPUSpec
}

type GPUSpec struct {
	HostRef              string
	Info                 GPUInfo
	AppInstanceModuleRef AppInstanceModuleRef
}

type AppInstanceModuleRef struct {
	AppInstanceRef
	Module  string
	Replica int
}

func (obj GPU) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

func (obj *GPU) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

func (obj GPU) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

type GPURegistry struct {
	registry.Registry
}

func (r GPURegistry) GetGPUName(hostRef string, slot int) string {
	return fmt.Sprintf("%s-slot-%d", hostRef, slot)
}

func NewGPU() *GPU {
	gpu := new(GPU)
	gpu.Init(ApiVersion, core.KindGPU)
	return gpu
}

func NewGPURegistry() GPURegistry {
	return GPURegistry{
		Registry: registry.NewRegistry(newGVK(core.KindGPU), false),
	}
}
