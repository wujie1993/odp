// Code generated by codegen. DO NOT EDIT!!!

package runtime

import (
	"github.com/wujie1993/waves/pkg/orm/core"
)

// DeepCopyInto is auto generated by codegen, copy public fields into the *AppInstance
func (src AppInstance) DeepCopyInto(dst *AppInstance) error {
	return core.DeepCopy(src, dst)
}

// DeepCopy is auto generated by codegen, create and copy public fields into the new *AppInstance
func (src AppInstance) DeepCopy() *AppInstance {
	dst := new(AppInstance)
	src.DeepCopyInto(dst)
	return dst
}

// DeepCopyApiObject is auto generated by codegen, deep copy and return as ApiObject
func (src AppInstance) DeepCopyApiObject() core.ApiObject {
	return src.DeepCopy()
}

// DeepCopyInto is auto generated by codegen, copy public fields into the *Host
func (src Host) DeepCopyInto(dst *Host) error {
	return core.DeepCopy(src, dst)
}

// DeepCopy is auto generated by codegen, create and copy public fields into the new *Host
func (src Host) DeepCopy() *Host {
	dst := new(Host)
	src.DeepCopyInto(dst)
	return dst
}

// DeepCopyApiObject is auto generated by codegen, deep copy and return as ApiObject
func (src Host) DeepCopyApiObject() core.ApiObject {
	return src.DeepCopy()
}

// DeepCopyInto is auto generated by codegen, copy public fields into the *Job
func (src Job) DeepCopyInto(dst *Job) error {
	return core.DeepCopy(src, dst)
}

// DeepCopy is auto generated by codegen, create and copy public fields into the new *Job
func (src Job) DeepCopy() *Job {
	dst := new(Job)
	src.DeepCopyInto(dst)
	return dst
}

// DeepCopyApiObject is auto generated by codegen, deep copy and return as ApiObject
func (src Job) DeepCopyApiObject() core.ApiObject {
	return src.DeepCopy()
}
