package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/wujie1993/waves/pkg/client"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/tests"
)

func init() {
	tests.ServeHTTP()
}

func TestGetHosts(t *testing.T) {
	cli := client.NewClientSet(tests.ServiceEndpoint)
	result, err := cli.V1().Hosts().List(context.TODO())
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", result)
}

func TestCreateHost(t *testing.T) {
	cli := client.NewClientSet(tests.ServiceEndpoint)
	host := v1.NewHost()
	host.Metadata.Name = "host-192.168.21.31"
	host.Metadata.Annotations["ShortName"] = "主机31"
	host.Spec.SSH = v1.HostSSH{
		Host:     "192.168.21.31",
		User:     "root",
		Password: "*****",
		Port:     22,
	}
	result, err := cli.V1().Hosts().Create(context.TODO(), host)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", result)
}

func TestGetHostWithTimeout(t *testing.T) {
	cli := client.NewClientSet(tests.ServiceEndpoint)
	ctx, _ := context.WithTimeout(context.Background(), 200*time.Millisecond)
	result, err := cli.V1().Hosts().Get(ctx, "host-192.168.21.31")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", result)
}

func TestUpdateHost(t *testing.T) {
	cli := client.NewClientSet(tests.ServiceEndpoint)
	host := v1.NewHost()
	host.Metadata.Name = "host-192.168.21.31"
	host.Metadata.Annotations["ShortName"] = "主机31"
	host.Spec.SSH = v1.HostSSH{
		Host:     "192.168.21.31",
		User:     "root",
		Password: "*****",
		Port:     2222,
	}
	result, err := cli.V1().Hosts().Update(context.TODO(), host)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", result)
}

func TestDeleteHost(t *testing.T) {
	cli := client.NewClientSet(tests.ServiceEndpoint)
	result, err := cli.V1().Hosts().Delete(context.TODO(), "host-192.168.21.31")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", result)
}
