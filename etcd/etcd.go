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

func (e *Etcd) Get(key string, f func(k, v []byte)) {
	if f == nil {
		return
	}

	resp, err := e.kv.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		e.log.Error(err)
		return
	}

	for _, kv := range resp.Kvs {
		if kv == nil {
			continue
		}
		f(kv.Key, kv.Value)
	}

}

func (e *Etcd) Put(key, value string, ttl ...int64) {
	var rttl int64 = 0
	if len(ttl) > 0 {
		rttl = ttl[0]
	}

	if rttl > 0 {
		lgr, err := e.lease.Grant(context.TODO(), rttl)
		if err != nil {
			e.log.Error(err)
			return
		}

		_, err = e.kv.Put(context.TODO(), key, value, clientv3.WithLease(lgr.ID))
		if err != nil {
			e.log.Error(err)
		}

		return
	}
	_, err := e.kv.Put(context.TODO(), key, value)
	if err != nil {
		e.log.Error(err)
	}
}

func (e *Etcd) Delete(key string, oldDataCallback ...func(k, v []byte)) {
	var cb func(k, v []byte) = nil
	if len(oldDataCallback) > 0 {
		cb = oldDataCallback[0]
	}

	resp, err := e.kv.Delete(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		e.log.Error(err)
		return
	}

	if cb != nil && len(resp.PrevKvs) > 0 {
		kv := resp.PrevKvs[0]
		cb(kv.Key, kv.Value)
	}
}

func (e *Etcd) Watch(key string, putCallback, delCallback func(k, v []byte)) {
	if putCallback == nil && delCallback == nil {
		return
	}

	resp, err := e.kv.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		e.log.Error(err)
		return
	}

	if putCallback != nil {
		for _, kvPair := range resp.Kvs {
			if kvPair == nil {
				continue
			}
			putCallback(kvPair.Key, kvPair.Value)
		}
	}

	go func(key string) {
		watchStartRevision := resp.Header.Revision + 1
		watchChan := e.watcher.Watch(context.TODO(), key,
			clientv3.WithPrefix(), clientv3.WithRev(watchStartRevision))

		for {
			select {
			case watchResp := <-watchChan:
				if watchResp.Events != nil {
					for _, watchEvent := range watchResp.Events {
						if watchEvent == nil {
							continue
						}
						we := watchEvent
						switch we.Type {
						case mvccpb.PUT:
							if putCallback != nil {
								putCallback(we.Kv.Key, we.Kv.Value)
							}
						case mvccpb.DELETE:
							if delCallback != nil {
								delCallback(we.Kv.Key, we.Kv.Value)
							}
						}
					}
				}
			}
		}
	}(key)

}

func (e *Etcd) KeepOnlineEx(key, value string, ttl int64) {
	if ttl <= 0 {
		return
	}

	var (
		err            error
		leaseGrantResp *clientv3.LeaseGrantResponse
		keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp  *clientv3.LeaseKeepAliveResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
	)

	for {
		leaseGrantResp, err = e.lease.Grant(context.TODO(), ttl)
		if err != nil {
			e.log.Error(err)
			goto RETRY
		}

		cancelFunc = nil

		keepAliveChan, err = e.lease.KeepAlive(context.TODO(), leaseGrantResp.ID)
		if err != nil {
			e.log.Error(err)
			goto RETRY
		}

		cancelCtx, cancelFunc = context.WithCancel(context.TODO())
		_, err = e.kv.Put(cancelCtx, key, value, clientv3.WithLease(leaseGrantResp.ID))
		if err != nil {
			e.log.Error(err)
			goto RETRY
		}

		for {
			select {
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil {
					goto RETRY
				}
			}
		}

	RETRY:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}

}

func (e *Etcd) KeepOnline(ttl int64, callback func()(string, string, bool)) {
	if callback == nil {
		e.log.Error("callback is nil")
		return
	}
	go func() {
		for {
			key, value, ok := callback()
			if !ok {
				goto RETRY
			}
			e.Put(key, value, ttl)
		RETRY:
			time.Sleep(time.Second * 1)
		}
	}()
}