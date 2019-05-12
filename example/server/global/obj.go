/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package global

const (
	GateKey = "helloGate"
	GameKey = "helloGame"
)

type Authorized struct {
	Type int32  `json:"type"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

type SessionData struct {
	SessionId string `json:"sessionId"`
}

type MsgTestReq struct {
	N int
}

type MsgTestAck struct {
	Str string
}
