package core

import (
	"time"
)

type BaseApiObj struct {
	MetaType `json:",inline" yaml:",inline"`
	BaseObj  `json:",inline" yaml:",inline"`
}

type ApiObjectAction struct {
	Type string
	Obj  ApiObject
}

type BaseRuntimeObj struct {
	BaseObj `json:",inline" yaml:",inline"`
}

type BaseObj struct {
	Metadata Metadata
	Status   Status
}

type Condition struct {
	Type               string
	Status             string
	LastTransitionTime time.Time
}

func (o BaseApiObj) GetKey() string {
	var key string
	if o.Metadata.Namespace != "" {
		key += "/namespaces/" + o.Metadata.Namespace
	}
	key += "/" + o.MetaType.Kind + "s/" + o.Metadata.Name
	return key
}

func (o BaseApiObj) GetGVK() GVK {
	return GVK{
		Group:      Group,
		ApiVersion: o.MetaType.ApiVersion,
		Kind:       o.MetaType.Kind,
	}
}

func (o *BaseApiObj) SetGVK(gvk GVK) {
	o.MetaType.ApiVersion = gvk.ApiVersion
	o.MetaType.Kind = gvk.Kind
}

func (o BaseApiObj) GetMetaType() MetaType {
	return o.MetaType
}

func (o BaseObj) GetMetadata() Metadata {
	return o.Metadata
}

func (o *BaseObj) SetMetadata(m Metadata) {
	o.Metadata = Metadata{}
	DeepCopy(&m, &o.Metadata)
}

func (o *BaseObj) SetNamespace(namespace string) {
	o.Metadata.Namespace = namespace
}

func (o *BaseObj) SetName(name string) {
	o.Metadata.Name = name
}

func (o *BaseObj) SetCreateTime(time time.Time) {
	o.Metadata.CreateTime = time
}

func (o *BaseObj) SetUpdateTime(time time.Time) {
	o.Metadata.UpdateTime = time
}

func (o *BaseObj) GetStatus() Status {
	return o.Status
}

func (o *BaseObj) SetStatus(status Status) {
	o.Status = Status{}
	DeepCopy(&status, &o.Status)
}

func (o *BaseObj) ResetConditions() {
	o.Status.Conditions = []Condition{}
}

func (o *BaseObj) SetStatusPhase(phase string) {
	o.Status.Phase = phase
}

func (o BaseObj) GetStatusPhase() string {
	return o.Status.Phase
}

func (o *BaseObj) RaiseVersion() {
	o.Metadata.ResourceVersion++
}

func (o *BaseApiObj) Init(apiVersion string, kind string) {
	o.ApiVersion = apiVersion
	o.Kind = kind
	o.BaseObj.Init()
}

// Init 初始化基础对象
func (o *BaseObj) Init() {
	o.Metadata.Init()
	o.Status = NewStatus()
}
