/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package common

import (
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/example/cluster/msg"
)

func GetAuthorizedData(typ int32, name, key string) []byte {
	d := msg.Authorized{
		Type: typ,
		Name: name,
		Key:  key,
	}
	return MarshalDataEx(d)
}

func AuthorizedCallback(conn defs.IConnection, packet defs.IPacket) (bool, int) {
	d := &msg.Authorized{}
	err := Unmarshal(packet.GetData(), d)
	if err != nil {
		return false, msg.RESULT_INVALID
	}

	ok := true
	srvType := int(d.Type)
	if srvType <= ST_NIL || srvType >= ST_MAX {
		conn.Close()
		ok = false

	} else {
		key := getAuthorizedKey(srvType)
		if d.Key != key {
			conn.Close()
			ok = false
		}
	}

	return ok, srvType
}

func getAuthorizedKey(srvType int) string {
	switch srvType {
	case ST_GATE:
		return GateKey
	case ST_LOGIC:
		return LogicKey
	}
	return ""
}
