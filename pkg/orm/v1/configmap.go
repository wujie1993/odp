package v1

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
)

type ConfigMap struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Data            map[string]string
}

func (obj ConfigMap) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Data)
}

func (obj *ConfigMap) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Data)
}

func (obj ConfigMap) SpecHash() string {
	data, _ := json.Marshal(&obj.Data)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

type ConfigMapRegistry struct {
	registry.Registry
}

func NewConfigMap() *ConfigMap {
	configMap := new(ConfigMap)
	configMap.Init(ApiVersion, core.KindConfigMap)
	configMap.Data = make(map[string]string)
	return configMap
}

func NewConfigMapRegistry() ConfigMapRegistry {
	return ConfigMapRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindConfigMap), true),
	}
}
