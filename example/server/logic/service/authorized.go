/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package service

import (
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/example/server/global"
)

func AuthorizedCallback(conn defs.IConnection, packet defs.IPacket) (bool, int) {
	d := &global.Authorized{}
	err := global.GetJSONMgr().Unmarshal(packet.GetData(), d)
	if err != nil {
		return false, global.Invalid
	}

	srvType := int(d.Type)
	switch srvType {
	case global.ST_GATE:
		if d.Key != global.GateKey {
			conn.Close()
			return false, global.Invalid
		}
	default:
		conn.Close()
	}

	return true, srvType
}
