package db_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/wujie1993/waves/pkg/db"
	_ "github.com/wujie1993/waves/tests"
)

const RegistryPrefix = "/registry-test"

func TestCRUD(t *testing.T) {
	key := RegistryPrefix + "/host/host1"
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
	if result, err := db.KV.List(RegistryPrefix+"/", true); err != nil {
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	watcher := db.KV.Watch(ctx, RegistryPrefix, true)
	received := false
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if watcher == nil {
			return
		}
		for {
			select {
			case item, ok := <-watcher:
				if !ok {
					return
				}
				t.Log(item)
				received = true
			}
		}
	}()
	go func() {
		defer wg.Done()
		TestCRUD(t)
	}()
	wg.Wait()
	if !received {
		t.Fatal("watch channel has nothing received")
	}
}

func TestRange(t *testing.T) {
	dataset := make(map[string]string)
	dataset[RegistryPrefix+"/audits/1587092947"] = "1"
	dataset[RegistryPrefix+"/audits/1587092952"] = "2"
	dataset[RegistryPrefix+"/audits/1587092960"] = "3"

	for k, v := range dataset {
		if err := db.KV.Set(k, v); err != nil {
			t.Fatal(err)
		}
	}

	result, err := db.KV.Range(RegistryPrefix+"/audits/1587092947", RegistryPrefix+"/audits/1587092952")
	if err != nil {
		t.Fatal(err)
		return
	}

	t.Logf("%+v", result)
}

func TestMutex(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := db.KV.Lock(ctx, RegistryPrefix+"/lock/lock1"); err != nil {
		t.Fatal(err)
	}
	if err := db.KV.Unlock(ctx, RegistryPrefix+"/lock/lock1"); err != nil {
		t.Fatal(err)
	}
}
