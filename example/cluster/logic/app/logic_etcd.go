/**
 * Created: 2020/3/26
 * @author: Jason
 */

package app

import (
	"fmt"
	"time"

	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/etcd"
	"github.com/lightning-go/lightning/example/cluster/common"
	"github.com/lightning-go/lightning/example/cluster/core"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/selector"
	"github.com/lightning-go/lightning/utils"
)

func (ls *LogicServer) initEtcd() {
	ls.registerEtcd()
}

func (ls *LogicServer) registerEtcd() bool {
	etcdCfg := conf.GetSrvCfg("etcd")
	if etcdCfg == nil {
		logger.Error("etcd config error")
		return false
	}

	ls.etcdMgr = etcd.NewEtcd(etcdCfg.HostList, time.Duration(etcdCfg.Timeout))
	if ls.etcdMgr == nil {
		logger.Error("create etcd failed")
		return false
	}

	ls.keepOnline()
	return true
}

func (ls *LogicServer) keepOnline() {
	if ls.etcdMgr == nil {
		return
	}

	cfg := ls.GetCfg()
	group := utils.IF(cfg != nil, cfg.Group, "group").(string)

	key := fmt.Sprintf("%v/%v/%v/%v", conf.GetServerName(), common.ETCD_LOGIC_PATH, group, ls.Name())
	name := ls.Name()
	host := ls.Host()

	go func() {
		for {
			weight := int(core.GetClientCount())
			sd := &selector.SessionData{
				Host:   host,
				Name:   name,
				Type:   common.ST_LOGIC,
				Weight: weight,
			}
			d, err := common.MarshalData(sd)
			if err != nil {
				logger.Error(err)
				goto RETRY
			}

			ls.etcdMgr.Put(key, string(d), 5)

		RETRY:
			time.Sleep(time.Second * 3)
		}
	}()

}
