/**
 * @author: Jason
 * Created: 19-5-6
 */

package etcd

import (
	"github.com/coreos/etcd/clientv3"
	"time"
	"github.com/lightning-go/lightning/logger"
	"context"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

const etcdLogPath = "./logs/etcd.log"

type Etcd struct {
	host        []string
	dialTimeout time.Duration
	client      *clientv3.Client
	kv          clientv3.KV
	lease       clientv3.Lease
	watcher     clientv3.Watcher
	log         *logger.Logger
}

func NewEtcd(host []string, timeout ...time.Duration) *Etcd {
	var t time.Duration = 0
	if len(timeout) > 0 {
		t = timeout[0]
	}
	e := &Etcd{
		host:        host,
		dialTimeout: t,
		log:         logger.NewLogger(logger.TRACE),
	}
	if e.log == nil {
		return nil
	}
	e.log.SetRotation(time.Hour*24*30, time.Hour*24, etcdLogPath)

	if !e.init() {
		return nil
	}

	return e
}

func (e *Etcd) init() bool {
	cfg := clientv3.Config{
		Endpoints:   e.host,
		DialTimeout: e.dialTimeout,
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		e.log.Error(err)
		return false
	}

	kv := clientv3.NewKV(client)
	if kv == nil {
		e.log.Error("init etcd kv failed")
		return false
	}

	lease := clientv3.NewLease(client)
	if lease == nil {
		e.log.Error("init etcd lease failed")
		return false
	}

	watcher := clientv3.NewWatcher(client)
	if watcher == nil {
		e.log.Error("init etcd watcher failed")
		return false
	}

	e.client = client
	e.kv = kv
	e.lease = lease
	e.watcher = watcher
	return true
}

func (e *Etcd) Watch(key string, putCallback, delCallback func(key, data []byte)) {
	if putCallback == nil {
		return
	}

	resp, err := e.kv.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		e.log.Error(err)
		return
	}

	for _, kvPair := range resp.Kvs {
		if kvPair == nil {
			continue
		}
		putCallback(kvPair.Key, kvPair.Value)
	}

	go func(key string) {
		watchStartRevision := resp.Header.Revision + 1
		watchChan := e.watcher.Watch(context.TODO(), key,
			clientv3.WithPrefix(), clientv3.WithRev(watchStartRevision))

		for {
			select {
			case watchResp := <- watchChan:
				if watchResp.Events == nil {
					break
				}
				for _, watchEvent := range watchResp.Events {
					if watchEvent == nil {
						continue
					}
					we := watchEvent
					switch we.Type {
					case mvccpb.PUT:
						putCallback(we.Kv.Key, we.Kv.Value)
					case mvccpb.DELETE:
						if delCallback != nil {
							putCallback(we.Kv.Key, we.Kv.Value)
						}
					}
				}
			}
		}
	}(key)

}
