/**
 * Created: 2020/3/26
 * @author: Jason
 */

package service

import (
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/example/cluster/msg"
)

type logicServe interface {
	SendDataToCenter(defs.ISession, interface{})
}

type logicService struct {
	serve logicServe
}

func NewLogicService(serve logicServe) *logicService {
	return &logicService{
		serve: serve,
	}
}

func (ls *logicService) Test(session defs.ISession, req *msg.MsgTestReq, ack *msg.MsgTestAck) int {
	ack.Result = req.N * 1000
	return msg.RESULT_OK
}

func (ls *logicService) Test2(session defs.ISession, req *msg.MsgTest2Req, ack *msg.MsgTest2Ack) int {
	ack.Result = req.N * 1000
	return msg.RESULT_OK
}

func (ls *logicService) Test3(session defs.ISession, req *msg.MsgTestCenterReq) int {
	ls.serve.SendDataToCenter(session, req)
	return msg.RESULT_OK
}
