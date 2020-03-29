/**
 * Created: 2020/3/27
 * @author: Jason
 */

package service

import (
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/example/cluster/msg"
)

type CenterService struct {

}


func (cs *CenterService) Test3(session defs.ISession, req *msg.MsgTestCenterAck, ack *msg.MsgTestCenterAck) int {
	ack.Result = req.Result + 200
	return msg.RESULT_OK
}
