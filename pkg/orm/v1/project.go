package v1

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
)

type Project struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	ReferNamespaces []string
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

func projectMutate(obj core.ApiObject) error {
	project := obj.(*Project)

	// 如果项目没有关联同名的命名空间，则创建并关联命名空间
	nsExist := false
	for _, referNs := range project.ReferNamespaces {
		if referNs == project.Metadata.Name {
			nsExist = true
		}
	}
	if !nsExist {
		nsRegistry := NewNamespaceRegistry()
		nsObj, err := nsRegistry.Get(context.TODO(), "", project.Metadata.Name)
		if err != nil {
			return err
		}
		if nsObj == nil {
			ns := NewNamespace()
			ns.Metadata.Name = project.Metadata.Name
			if _, err := nsRegistry.Create(context.TODO(), ns); err != nil {
				return err
			}
		}
		project.ReferNamespaces = append(project.ReferNamespaces, project.Metadata.Name)
	}
	return nil
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
	r := ProjectRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindProject), false),
	}
	r.SetMutateHook(projectMutate)
	return r
}
