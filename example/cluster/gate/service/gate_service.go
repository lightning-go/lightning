/**
 * Created: 2020/3/25
 * @author: Jason
 */

package service

import (
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/example/cluster/msg"
	"time"
)

type GateService struct {}

func (gs *GateService) Ping(session defs.ISession, req *msg.MsgPingReq, ack *msg.MsgPingAck) int {
	t := time.Now()
	now := t.Unix()
	zone, offset := t.Zone()
	ack.Now = now
	ack.Zone = zone
	ack.Offset = offset
	return msg.RESULT_OK
}

func (gs *GateService) TestNil(session defs.ISession, req *msg.MsgTestNilReq, ack *msg.MsgTestNilAck) int {
	ack.Res = "hello nil"
	return msg.RESULT_OK
}
