package rest_test

import (
	"context"
	"testing"

	"github.com/wujie1993/waves/pkg/client/rest"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/tests"
)

func TestGetHosts(t *testing.T) {
	restCli := rest.NewRESTClient(tests.ServiceEndpoint)
	hosts := &[]v1.Host{}
	if err := restCli.Get().Version("v1").Resource("hosts").Do(context.TODO()).Into(hosts); err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", hosts)
}

func TestGetAppInstances(t *testing.T) {
	restCli := rest.NewRESTClient(tests.ServiceEndpoint)
	appInstances := &[]v1.AppInstance{}
	if err := restCli.Get().Version("v1").Namespace("default").Resource("appinstances").Do(context.TODO()).Into(appInstances); err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", appInstances)
}
