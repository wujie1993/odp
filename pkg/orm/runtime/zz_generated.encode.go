// Code generated by codegen. DO NOT EDIT!!!

package runtime

import (
	"encoding/json"

	"github.com/ghodss/yaml"
)

// ToJSON is auto generated by codegen, marshal to json bytes
func (obj AppInstance) ToJSON() ([]byte, error) {
	return json.Marshal(obj)
}

// ToJSONPretty is auto generated by codegen, marshal to json bytes with pretty format
func (obj AppInstance) ToJSONPretty() ([]byte, error) {
	return json.MarshalIndent(obj, "", "\t")
}

// FromJSON is auto generated by codegen, unmarshal from json bytes
func (obj *AppInstance) FromJSON(data []byte) error {
	return json.Unmarshal(data, obj)
}

// ToYAML is auto generated by codegen, marshal to yaml bytes
func (obj AppInstance) ToYAML() ([]byte, error) {
	return yaml.Marshal(obj)
}

// FromYAML is auto generated by codegen, unmarshal from yaml bytes
func (obj *AppInstance) FromYAML(data []byte) error {
	return yaml.Unmarshal(data, obj)
}

// ToJSON is auto generated by codegen, marshal to json bytes
func (obj Host) ToJSON() ([]byte, error) {
	return json.Marshal(obj)
}

// ToJSONPretty is auto generated by codegen, marshal to json bytes with pretty format
func (obj Host) ToJSONPretty() ([]byte, error) {
	return json.MarshalIndent(obj, "", "\t")
}

// FromJSON is auto generated by codegen, unmarshal from json bytes
func (obj *Host) FromJSON(data []byte) error {
	return json.Unmarshal(data, obj)
}

// ToYAML is auto generated by codegen, marshal to yaml bytes
func (obj Host) ToYAML() ([]byte, error) {
	return yaml.Marshal(obj)
}

// FromYAML is auto generated by codegen, unmarshal from yaml bytes
func (obj *Host) FromYAML(data []byte) error {
	return yaml.Unmarshal(data, obj)
}

// ToJSON is auto generated by codegen, marshal to json bytes
func (obj Job) ToJSON() ([]byte, error) {
	return json.Marshal(obj)
}

// ToJSONPretty is auto generated by codegen, marshal to json bytes with pretty format
func (obj Job) ToJSONPretty() ([]byte, error) {
	return json.MarshalIndent(obj, "", "\t")
}

// FromJSON is auto generated by codegen, unmarshal from json bytes
func (obj *Job) FromJSON(data []byte) error {
	return json.Unmarshal(data, obj)
}

// ToYAML is auto generated by codegen, marshal to yaml bytes
func (obj Job) ToYAML() ([]byte, error) {
	return yaml.Marshal(obj)
}

// FromYAML is auto generated by codegen, unmarshal from yaml bytes
func (obj *Job) FromYAML(data []byte) error {
	return yaml.Unmarshal(data, obj)
}
