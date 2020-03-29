/**
 * Created: 2020/3/27
 * @author: Jason
 */

package service

import (
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/example/cluster/msg"
)

type LogicService struct {}


func (ls *LogicService) Test3(session defs.ISession, req *msg.MsgTestCenterReq, ack *msg.MsgTestCenterAck) int {
	ack.Result = req.N * 2000
	return msg.RESULT_OK
}
