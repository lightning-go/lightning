/**
 * Created: 2020/3/25
 * @author: Jason
 */

package msg

type Authorized struct {
	Type int32  `json:"type"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

type SessionData struct {
	SessionId string `json:"sessionId"`
}

type MsgPingReq struct {
}

type MsgPingAck struct {
	Now    int64  `json:"now"`
	Zone   string `json:"zone"`
	Offset int    `json:"offset"`
}

type MsgTestReq struct {
	N int64 `json:"n"`
}

type MsgTestAck struct {
	Result int64 `json:"result"`
}

type MsgTest2Req struct {
	N int64 `json:"n"`
}

type MsgTest2Ack struct {
	Result int64 `json:"result"`
	Name   string
}

type MsgTestCenterReq struct {
	N int64 `json:"n"`
}

type MsgTestCenterAck struct {
	Result int64 `json:"result"`
}
