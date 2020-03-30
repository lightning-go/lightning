/**
 * Created: 2020/3/25
 * @author: Jason
 */

package app

import (
	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/etcd"
	"time"
	"fmt"
	"github.com/lightning-go/lightning/example/cluster/common"
	"github.com/lightning-go/lightning/selector"
	"strings"
)

func (gs *GateServer) watch() bool {
	etcdCfg := conf.GetServer("etcd")
	if etcdCfg == nil {
		logger.Error("etcd config error")
		return false
	}

	srvCfg := gs.GetCfg()
	if srvCfg == nil {
		logger.Errorf("%v config load failed", gs.Name())
		return false
	}

	gs.etcdMgr = etcd.NewEtcd(etcdCfg.HostList, time.Duration(etcdCfg.Timeout))
	if gs.etcdMgr == nil {
		logger.Error("create etcd failed")
		return false
	}

	key := fmt.Sprintf("%v/%v", common.ETCD_LOGIC_PATH, srvCfg.WatchGroup)
	gs.etcdMgr.Watch(key, func(k, v []byte) {
		sd := &selector.SessionData{}
		err := common.Unmarshal(v, sd)
		if err != nil {
			logger.Error(err)
			return
		}

		gs.serveSelector.AddRemoteData(sd, func(data *selector.SessionData) bool {
			return gs.initRemote(data)
		})

		logger.Debugf("watch add data: %s, %v", k, sd)

	}, func(k, v []byte) {
		logger.Tracef("watch delete data: %s", k)

		key := string(k)
		idx := strings.LastIndex(key, "/")
		if idx == -1 {
			logger.Errorf("key %v error", key)
			return
		}
		key = key[idx + 1:]
		gs.serveSelector.DelRemoteData(key)

		logger.Debugf("watch delete key: %s", key)

	})

	return true
}

