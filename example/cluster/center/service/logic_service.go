/**
 * Created: 2020/3/27
 * @author: Jason
 */

package service

import (
	"github.com/lightning-go/lightning/example/cluster/msg"
	"github.com/lightning-go/lightning/example/cluster/data"
)

type LogicService struct {}


func (ls *LogicService) Test3(client *data.Client, req *msg.MsgTestCenterReq, ack *msg.MsgTestCenterAck) int {
	ack.Result = req.N * 2000
	return msg.RESULT_OK
}
