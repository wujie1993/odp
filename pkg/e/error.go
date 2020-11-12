package e

import (
	"errors"
	"fmt"
)

func Errorf(format string, args ...interface{}) error {
	return errors.New(fmt.Sprintf(format, args...))
}
