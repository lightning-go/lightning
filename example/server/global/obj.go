/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package global

const (
	GateKey = "helloGate"
)

type Authorized struct {
	Type int32  `json:"type"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

type MsgTestReq struct {
	N int
}

type MsgTestAck struct {
	Str string
}
