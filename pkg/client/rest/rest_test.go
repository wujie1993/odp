package rest_test

import (
	"testing"

	"github.com/wujie1993/waves/pkg/client/rest"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

func TestGetHosts(t *testing.T) {
	restCli := rest.NewRESTClient("http://172.25.23.200/deployer")
	hosts := &[]v1.Host{}
	if err := restCli.Get().Version("v1").Resource("hosts").Do().Into(hosts); err != nil {
		t.Error(err)
		return
	}
	//t.Logf("%+v", hosts)
}

func TestGetAppInstances(t *testing.T) {
	restCli := rest.NewRESTClient("http://172.25.23.200/deployer")
	appInstances := &[]v1.AppInstance{}
	if err := restCli.Get().Version("v1").Namespace("default").Resource("appinstances").Do().Into(appInstances); err != nil {
		t.Error(err)
		return
	}
	//t.Logf("%+v", appInstances)
}

func TestGetEvents(t *testing.T) {
	restCli := rest.NewRESTClient("http://172.25.23.200/deployer")
	events := &[]v1.Event{}
	if err := restCli.Get().Version("v1").
		Resource("events").
		Params(map[string]string{"resourceKind": "appInstance", "action": "HealthCheck"}).
		Do().Into(events); err != nil {
		t.Error(err)
		return
	}
	//t.Logf("%+v", events)
}
