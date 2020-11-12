package util_test

import (
	"testing"

	"github.com/wujie1993/waves/pkg/util"
)

func TestInstancePool(t *testing.T) {
	dataDir := "/tmp/instance_pool"
	id, err := util.AddInstancePoolItem(dataDir, "node-32", "videoanalysis-apps", "9e0ee849-aba2-417f-be3a-e1d1c4555218")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(id)
}
