package core

import (
	"time"
)

type Metadata struct {
	Name            string
	Namespace       string `json:",omitempty" yaml:",omitempty"`
	Uid             string
	Labels          map[string]string
	Annotations     map[string]string
	ResourceVersion int
	CreateTime      time.Time
	UpdateTime      time.Time
	Finalizers      []string
}

type MetaType struct {
	Kind       string
	ApiVersion string
}

func (m *Metadata) Init() {
	m.Labels = make(map[string]string)
	m.Annotations = make(map[string]string)
	m.Finalizers = []string{}
}

func (m Metadata) CopyTo(dest *Metadata) {
	DeepCopy(&m, &dest)
}

func (m Metadata) GetKey(kind string, namespaced bool) string {
	var key string
	if namespaced {
		if m.Namespace != "" {
			key += "/namespaces/" + m.Namespace
		} else {
			key += "/namespaces/" + DefaultNamespace
		}
	}
	return key + "/" + GetPlural(kind) + "/" + m.Name
}
