package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/wujie1993/waves/pkg/client"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

func TestGetHosts(t *testing.T) {
	cli := client.NewClientSet("http://localhost:8000/deployer")
	result, err := cli.V1().Hosts().List(context.TODO())
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", result)
}

func TestGetHostWithTimeout(t *testing.T) {
	cli := client.NewClientSet("http://localhost:8000/deployer")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Millisecond)
	result, err := cli.V1().Hosts().Get(ctx, "host-172.25.21.32")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", result)
}

func TestCreateHost(t *testing.T) {
	cli := client.NewClientSet("http://localhost:8000/deployer")
	host := v1.NewHost()
	host.Metadata.Name = "host-172.25.21.32"
	host.Metadata.Annotations["ShortName"] = "主机32"
	host.Spec.SSH = v1.HostSSH{
		Host:     "172.25.21.32",
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

func TestUpdateHost(t *testing.T) {
	cli := client.NewClientSet("http://localhost:8000/deployer")
	host := v1.NewHost()
	host.Metadata.Name = "host-172.25.21.32"
	host.Metadata.Annotations["ShortName"] = "主机32"
	host.Spec.SSH = v1.HostSSH{
		Host:     "172.25.21.32",
		User:     "root",
		Password: "*****",
		Port:     22,
	}
	result, err := cli.V1().Hosts().Update(context.TODO(), host)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", result)
}

func TestDeleteHost(t *testing.T) {
	cli := client.NewClientSet("http://localhost:8000/deployer")
	result, err := cli.V1().Hosts().Delete(context.TODO(), "host-172.25.21.32")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", result)
}

func TestGetAppInstances(t *testing.T) {
	cli := client.NewClientSet("http://localhost:8000/deployer")
	result, err := cli.V2().AppInstances("default").List(context.TODO())
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", result)
}

func TestGetAppInstance(t *testing.T) {
	cli := client.NewClientSet("http://localhost:8000/deployer")
	result, err := cli.V2().AppInstances("default").Get(context.TODO(), "mysql-cb5bcbff")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", result)
}
