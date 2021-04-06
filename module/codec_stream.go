/**
 * Created: 2019/4/20 0020
 * @author: Jason
 */

package module

import (
	"encoding/binary"
	"github.com/lightning-go/lightning/defs"
)

type StreamCodec struct {
	dec  *Decoder
	enc  *Encoder
	buff []byte
}

func NewStreamCodec() *StreamCodec {
	return &StreamCodec{}
}

func (sc *StreamCodec) Init(conn defs.IConnection) bool {
	if conn == nil {
		return false
	}
	iTcpConn, ok := conn.(defs.ITcpConnection)
	if !ok {
		return false
	}
	c := iTcpConn.GetConn()
	if c == nil {
		return false
	}

	sc.dec = NewDecoder(c, binary.BigEndian)
	sc.enc = NewEncode(c, binary.BigEndian)

	sc.buff = make([]byte, DefaultBufferSize)
	return true
}
func (sc *StreamCodec) Write(packet defs.IPacket) error {
	if sc.enc == nil {
		return ErrCodecWriteNil
	}

	err := sc.enc.EncodeData(packet.GetData())
	if err != nil {
		return err
	}

	sc.enc.Flush()
	return nil
}

func (sc *StreamCodec) Read() (defs.IPacket, error) {
	if sc.dec == nil {
		return nil, ErrCodecReadNil
	}

	n, err := sc.dec.DecodeData(sc.buff)
	if err != nil {
		return nil, err
	}

	p := &defs.Packet{}
	p.SetData(sc.buff[:n])

	return p, nil
}
