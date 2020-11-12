package v1

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
)

type Project struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
}

func (obj Project) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj)
}

func (obj *Project) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj)
}

func (obj Project) SpecHash() string {
	data, _ := json.Marshal(&obj)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

type ProjectRegistry struct {
	registry.Registry
}

func NewProject() *Project {
	project := new(Project)
	project.Init(ApiVersion, core.KindProject)
	return project
}

func NewProjectRegistry() ProjectRegistry {
	return ProjectRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindProject), false),
	}
}
