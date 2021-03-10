/**
 * Created: 2020/3/27
 * @author: Jason
 */

package service

import (
	"github.com/lightning-go/lightning/example/cluster/msg"
	"github.com/lightning-go/lightning/example/cluster/core"
)

type CenterService struct {

}


func (cs *CenterService) Test3(client *core.Client, req *msg.MsgTestCenterAck, ack *msg.MsgTestCenterAck) int {
	ack.Result = req.Result + 200
	return msg.RESULT_OK
}
