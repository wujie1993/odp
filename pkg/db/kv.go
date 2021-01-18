package db

import (
	"context"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3/concurrency"

	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/setting"
)

const (
	KVActionTypeSet    = "set"
	KVActionTypeDelete = "delete"
)

var KV KVStorage
var ClientV3 *clientv3.Client

type KVStorage interface {
	Get(string) (string, error)
	Set(string, string) error
	List(string, bool) (map[string]string, error)
	Delete(string) (string, error)
	Range(string, string) (map[string]string, error)
	Watch(context.Context, string, bool) <-chan KVAction
	Lock(context.Context, string) error
	Unlock(context.Context, string) error
}

type KVAction struct {
	Key        string
	Value      string
	ActionType string
}

type EtcdClient struct {
	client     *clientv3.Client
	endpoints  string
	timeout    time.Duration
	retryTimes int

	kvMutexMap      map[string]KVMutex
	kvMutexMapMutex sync.Mutex
}

type KVMutex struct {
	session *concurrency.Session
	mutex   *concurrency.Mutex
}

func InitKV() error {
	etcdCli, err := NewEtcdClient(setting.EtcdSetting.Endpoints)
	if err != nil {
		return err
	}
	KV = etcdCli
	ClientV3 = etcdCli.client
	return nil
}

func NewEtcdClient(endpoints []string) (*EtcdClient, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	etcdClient := &EtcdClient{
		client:     cli,
		timeout:    time.Duration(5 * time.Second),
		retryTimes: 3,
		kvMutexMap: make(map[string]KVMutex),
	}
	return etcdClient, nil
}

func (c *EtcdClient) Get(key string) (string, error) {
	for retry := 0; retry < c.retryTimes; retry++ {
		ctx, _ := context.WithTimeout(context.Background(), c.timeout)
		resp, err := c.client.Get(ctx, key)
		if err != nil {
			switch err {
			case context.DeadlineExceeded:
				log.Warn(err)
				// 出现超时错误时进行重试
				continue
			default:
				log.Error(err)
				return "", err
			}
		}
		if resp.Count < 1 {
			return "", nil
		}
		valueBytes, err := base64.RawStdEncoding.DecodeString(string(resp.Kvs[0].Value))
		if err != nil {
			return "", err
		}
		return string(valueBytes), nil
	}
	return "", errors.New("failed to get " + key)
}

func (c *EtcdClient) Range(begin string, end string) (map[string]string, error) {
	for retry := 0; retry < c.retryTimes; retry++ {
		ctx, _ := context.WithTimeout(context.Background(), c.timeout)
		resp, err := c.client.Get(ctx, begin, clientv3.WithRange(end))
		if err != nil {
			switch err {
			case context.DeadlineExceeded:
				log.Warn(err)
				// 出现超时错误时进行重试
				continue
			default:
				log.Error(err)
				return nil, err
			}
		}
		result := make(map[string]string)
		for _, kvs := range resp.Kvs {
			valueBytes, err := base64.RawStdEncoding.DecodeString(string(kvs.Value))
			if err != nil {
				return nil, err
			}
			result[string(kvs.Key)] = string(valueBytes)
		}
		return result, nil
	}
	return nil, errors.New("failed to range " + begin + " to " + end)
}

func (c *EtcdClient) Set(key string, value string) error {
	base64Value := base64.RawStdEncoding.EncodeToString([]byte(value))
	for retry := 0; retry < c.retryTimes; retry++ {
		ctx, _ := context.WithTimeout(context.Background(), c.timeout)
		if _, err := c.client.Put(ctx, key, base64Value); err != nil {
			switch err {
			case context.DeadlineExceeded:
				log.Warn(err)
				// 出现超时错误时进行重试
				continue
			default:
				log.Error(err)
				return err
			}
		}
		return nil
	}
	return errors.New("failed to set " + key)
}

func (c *EtcdClient) Delete(key string) (string, error) {
	for retry := 0; retry < c.retryTimes; retry++ {
		ctx, _ := context.WithTimeout(context.Background(), c.timeout)
		resp, err := c.client.Delete(ctx, key, clientv3.WithPrevKV())
		if err != nil {
			switch err {
			case context.DeadlineExceeded:
				log.Warn(err)
				// 出现超时错误时进行重试
				continue
			default:
				log.Error(err)
				return "", err
			}
		}
		if len(resp.PrevKvs) < 1 {
			return "", nil
		}
		valueBytes, err := base64.RawStdEncoding.DecodeString(string(resp.PrevKvs[0].Value))
		if err != nil {
			return "", err
		}
		return string(valueBytes), nil
	}
	return "", errors.New("failed to delete " + key)
}

func (c *EtcdClient) List(key string, withPrefix bool) (map[string]string, error) {
	var err error
	var resp *clientv3.GetResponse

	for retry := 0; retry < c.retryTimes; retry++ {
		ctx, _ := context.WithTimeout(context.Background(), c.timeout)
		if withPrefix {
			resp, err = c.client.Get(ctx, key, clientv3.WithPrefix())
		} else {
			resp, err = c.client.Get(ctx, key)
		}
		if err != nil {
			switch err {
			case context.DeadlineExceeded:
				log.Warn(err)
				// 出现超时错误时进行重试
				continue
			default:
				log.Error(err)
				return nil, err
			}
		}
		result := make(map[string]string)
		for _, kvs := range resp.Kvs {
			valueBytes, err := base64.RawStdEncoding.DecodeString(string(kvs.Value))
			if err != nil {
				return nil, err
			}
			result[string(kvs.Key)] = string(valueBytes)
		}
		return result, nil
	}
	return nil, errors.New("failed to list " + key)
}

func (c *EtcdClient) Watch(ctx context.Context, key string, withPrefix bool) <-chan KVAction {
	if _, err := c.client.Dial(c.client.Endpoints()[0]); err != nil {
		return nil
	}

	var kvsWatcher clientv3.WatchChan
	if withPrefix {
		kvsWatcher = c.client.Watch(ctx, key, clientv3.WithPrefix(), clientv3.WithPrevKV())
	} else {
		kvsWatcher = c.client.Watch(ctx, key, clientv3.WithPrevKV())
	}
	watcher := make(chan KVAction, 1000)
	go func() {
		defer close(watcher)
		for {
			select {
			case <-ctx.Done():
				return
			case wresp, ok := <-kvsWatcher:
				if !ok {
					return
				}
				for _, event := range wresp.Events {
					switch event.Type {
					case mvccpb.PUT:
						valueBytes, err := base64.RawStdEncoding.DecodeString(string(event.Kv.Value))
						if err != nil {
							log.Error(err)
						}
						watcher <- KVAction{
							Key:        string(event.Kv.Key),
							Value:      string(valueBytes),
							ActionType: KVActionTypeSet,
						}
					case mvccpb.DELETE:
						if event.PrevKv == nil {
							continue
						}

						valueBytes, err := base64.RawStdEncoding.DecodeString(string(event.PrevKv.Value))
						if err != nil {
							log.Error(err)
						}
						watcher <- KVAction{
							Key:        string(event.PrevKv.Key),
							Value:      string(valueBytes),
							ActionType: KVActionTypeDelete,
						}
					default:
						log.Warn(errors.New("unknown kv action type"))
					}
				}
			}
		}
	}()
	return watcher
}

func (c *EtcdClient) Lock(ctx context.Context, key string) error {
	grantCtx, _ := context.WithCancel(ctx)
	lease, err := c.client.Grant(grantCtx, 30)
	if err != nil {
		log.Error(err)
		return err
	}
	ss, err := concurrency.NewSession(c.client, concurrency.WithLease(lease.ID))
	if err != nil {
		log.Error(err)
		ss.Close()
		return err
	}
	mtx := concurrency.NewMutex(ss, key)
	lockCtx, _ := context.WithCancel(ctx)
	if err := mtx.Lock(lockCtx); err != nil {
		log.Error(err)
		ss.Close()
		return err
	}
	c.kvMutexMapMutex.Lock()
	c.kvMutexMap[key] = KVMutex{
		session: ss,
		mutex:   mtx,
	}
	c.kvMutexMapMutex.Unlock()
	return nil
}

func (c *EtcdClient) Unlock(ctx context.Context, key string) error {
	c.kvMutexMapMutex.Lock()
	defer c.kvMutexMapMutex.Unlock()

	kvMutex, ok := c.kvMutexMap[key]
	if !ok {
		err := e.Errorf("lock key %s not found", key)
		log.Error(err)
		return err
	}
	kvMutex.session.Close()
	unlockCtx, _ := context.WithCancel(ctx)
	kvMutex.mutex.Unlock(unlockCtx)
	delete(c.kvMutexMap, key)
	return nil
}
