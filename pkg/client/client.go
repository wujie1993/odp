package client

import (
	"github.com/wujie1993/waves/pkg/client/rest"
	"github.com/wujie1993/waves/pkg/client/v1"
	"github.com/wujie1993/waves/pkg/client/v2"
)

type ClientSet struct {
	v1 v1.Client
	v2 v2.Client
}

func (s ClientSet) V1() v1.Client {
	return s.v1
}

func (s ClientSet) V2() v2.Client {
	return s.v2
}

func NewClientSet(endpoint string) ClientSet {
	return ClientSet{
		v1: v1.NewClient(rest.NewRESTClient(endpoint)),
		v2: v2.NewClient(rest.NewRESTClient(endpoint)),
	}
}
