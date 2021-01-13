package runtime

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/wujie1993/waves/pkg/orm/core"
)

func (obj Host) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

func (obj *Host) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

func (obj Host) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func NewHost() *Host {
	host := new(Host)
	host.Init("", core.KindHost)
	return host
}
