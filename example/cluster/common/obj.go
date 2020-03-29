/**
 * Created: 2020/3/25
 * @author: Jason
 */

package common

import (
	"github.com/json-iterator/go"
	"github.com/lightning-go/lightning/utils"
	"bytes"
	"compress/zlib"
	"io"
)

const (
	ST_NIL    = iota
	ST_GATE
	ST_LOGIC
	ST_CENTER
	ST_FIGHT
	ST_MAX
)

const (
	GateKey = "helloGateo3@#^34Sdsdfj@5"
	LogicKey = "helloLogic*&12@$sdlkl$sdf"
)

const (
	ETCD_LOGIC_PATH = "/game/logic"
	ETCD_GATE_PATH  = "/game/gate"
)

func MarshalDataEx(v interface{}) []byte {
	d, err := jsoniter.Marshal(v)
	if err != nil || d == nil {
		return utils.NullData
	}
	return d
}

func MarshalData(v interface{}) ([]byte, error) {
	return jsoniter.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return jsoniter.Unmarshal(data, v)
}

func ZlibCompress(src []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write(src)
	w.Close()
	return in.Bytes()
}

func ZlibUnCompress(src []byte) []byte {
	b := bytes.NewReader(src)
	var out bytes.Buffer
	r, _ := zlib.NewReader(b)
	io.Copy(&out, r)
	return out.Bytes()
}
