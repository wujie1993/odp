package core

import (
	"time"
)

type ApiObject interface {
	JSONMarshaler
	YAMLMarshaler

	MetadataGetter
	MetadataSetter
	MetaTypeGetter
	SpecHasher
	ConditionResetter
	Versioner

	GetKey() string
	SetNamespace(string)
	SetName(string)
	GetGVK() GVK
	SetGVK(gvk GVK)
	GetStatus() Status
	SetStatus(Status)
	SetStatusPhase(string)
	GetStatusPhase() string

	Sha256() string
	DeepCopyApiObject() ApiObject
}

type ApiObjectList []ApiObject

type JSONMarshaler interface {
	ToJSON() ([]byte, error)
	FromJSON([]byte) error
}

type YAMLMarshaler interface {
	ToYAML() ([]byte, error)
	FromYAML([]byte) error
}

type RuntimeObject interface {
	MetadataGetter
}

type MetadataGetter interface {
	GetMetadata() Metadata
}

type MetadataSetter interface {
	SetMetadata(Metadata)
	SetCreateTime(time.Time)
	SetUpdateTime(time.Time)
}

type MetaTypeGetter interface {
	GetMetaType() MetaType
}

type ConditionResetter interface {
	ResetConditions()
}

type SpecHasher interface {
	SpecEncode() ([]byte, error)
	SpecDecode([]byte) error
	SpecHash() string
}

type Versioner interface {
	RaiseVersion()
}
