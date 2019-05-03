/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package service

import (
	"github.com/lightning-go/lightning/example/server/global"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
)

type Service struct{}

func (s *Service) Test(conn defs.ISession, req *global.MsgTestReq, ack *global.MsgTestAck) int {
	logger.Tracef("%v", req.N)
	ack.Str = "hi..........."

	return global.RESULT_OK
}
