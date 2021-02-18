package e

import (
	"errors"
	"fmt"
)

func Errorf(format string, args ...interface{}) error {
	return errors.New(fmt.Sprintf(format, args...))
}

type ResourceExistsError struct {
	Key string
}

func (e ResourceExistsError) Error() string {
	return fmt.Sprintf("资源 %s 已存在", e.Key)
}

type ResourceNotFoundError struct {
	Key string
}

func (e ResourceNotFoundError) Error() string {
	return fmt.Sprintf("资源 %s 不存在", e.Key)
}

type OperationForbidenError struct{}

func (e OperationForbidenError) Error() string {
	return "禁止操作"
}

type InvalidNamespaceError struct {
	Namespace string
}

func (e InvalidNamespaceError) Error() string {
	return fmt.Sprintf("无效的命名空间 %s", e.Namespace)
}

type InvalidNameError struct {
	Name string
}

func (e InvalidNameError) Error() string {
	return fmt.Sprintf("无效的名称 %s", e.Name)
}

type JobExecTimeoutError struct{}

func (e JobExecTimeoutError) Error() string {
	return "任务运行超时"
}
