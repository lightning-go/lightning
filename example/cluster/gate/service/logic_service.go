/**
 * Created: 2020/3/26
 * @author: Jason
 */

package service

import (
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/example/cluster/msg"
)

type LogicService struct{}

func (ls *LogicService) Test2(session defs.ISession, ack *msg.MsgTest2Ack, ack2 *msg.MsgTest2Ack) int {
	ack2.Result = ack.Result * 1000
	ack2.Name = "test..."

	return msg.RESULT_OK
}

