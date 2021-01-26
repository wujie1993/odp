package loader_test

import (
	"testing"

	"github.com/wujie1993/waves/pkg/wavectl/loader"
)

func TestGetHosts(t *testing.T) {
	objs, err := loader.LoadObjsByLocalPath("../../../tests/host-test.yaml")
	if err != nil {
		t.Error(err)
		return
	}
	for _, obj := range objs {
		data, _ := obj.ToYAML()
		t.Log(string(data))
	}
}
