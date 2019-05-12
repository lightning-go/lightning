/**
 * @author: Jason
 * Created: 19-5-12
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
	case global.ST_GAME:
		if d.Key != global.GameKey {
			conn.Close()
			return false, global.Invalid
		}
	default:
		conn.Close()
	}

	return true, srvType
}


type LogicService struct {}


