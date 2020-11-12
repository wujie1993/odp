package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/setting"
)

func init() {
	setting.EtcdSetting = &setting.Etcd{
		Endpoints: []string{"localhost:2379"},
	}
	db.InitKV()
}

func TestCRUD(t *testing.T) {
	key := "/prophet/host/host1"
	value := "{\"name\":\"host1\",\"address\":\"192.168.1.1\"}"

	// 写入
	if err := db.KV.Set(key, value); err != nil {
		t.Fatal(err)
	}

	// 读取
	respValue, err := db.KV.Get(key)
	if err != nil {
		t.Fatal(err)
	} else if respValue != value {
		t.Fatal("get result incorrect")
	}

	value = "{\"name\":\"host1\",\"address\":\"172.21.1.1\"}"
	// 更新
	if err := db.KV.Set(key, value); err != nil {
		t.Fatal(err)
	}

	// 列举
	if result, err := db.KV.List("/prophet/", true); err != nil {
		t.Fatal(err)
	} else if result[key] != value {
		t.Fatal("list result incorrect")
	} else {
		t.Logf("list: %v+", result)
	}

	// 删除
	if result, err := db.KV.Delete(key); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("delete %s", result)
	}

	if respValue, err = db.KV.Get(key); err != nil {
		t.Fatal(err)
	} else if respValue != "" {
		t.Fatal("delete unsuccessful")
	}
}

func TestWatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	watcher := db.KV.Watch(ctx, "/prophet", true)
	go func() {
		for item := range watcher {
			t.Log(item)
		}
	}()
	go func() {
		TestCRUD(t)
	}()
	time.Sleep(5 * time.Second)
}

func TestRange(t *testing.T) {
	dataset := make(map[string]string)
	dataset["/audits/1587092947"] = "1"
	dataset["/audits/1587092952"] = "2"
	dataset["/audits/1587092960"] = "3"

	for k, v := range dataset {
		if err := db.KV.Set(k, v); err != nil {
			t.Error(err)
		}
	}

	result, err := db.KV.Range("/audits/1587092947", "/audits/1587092952")
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", result)
}

func TestMutex(t *testing.T) {
	ctx := context.Background()
	if err := db.KV.Lock(ctx, "/lock/lock1"); err != nil {
		t.Error(err)
	}
	if err := db.KV.Unlock(ctx, "/lock/lock1"); err != nil {
		t.Error(err)
	}
}
