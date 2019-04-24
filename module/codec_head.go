/**
 * Created: 2019/4/22 0022
 * @author: Jason
 */

package module

import (
	"github.com/lightning-go/lightning/defs"
)

type HeadCodec struct {
	dec  *Decoder
	enc  *Encoder
}

func NewHeadCodec() *HeadCodec {
	return &HeadCodec{}
}

func (hc *HeadCodec) Init(conn defs.IConnection) bool {
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

	hc.dec = NewDecoder(c)
	hc.enc = NewEocode(c)

	return true
}
func (hc *HeadCodec) Write(packet defs.IPacket) error {
	if hc.enc == nil {
		return ErrCodecWriteNil
	}

	data := packet.GetData()
	dataLen := len(data)
	err := hc.enc.EncodeInt32(int32(dataLen))
	if err != nil {
		hc.enc.Clean()
		return err
	}

	err = hc.enc.EncodeData(data)
	if err != nil {
		hc.enc.Clean()
		return err
	}

	hc.enc.Flush()
	return nil
}

func (hc *HeadCodec) Read() (defs.IPacket, error) {
	if hc.dec == nil {
		return nil, ErrCodecReadNil
	}

	dataLen, err := hc.dec.DecodeInt32()
	if err != nil {
		hc.dec.Clean()
		return nil, err
	}

	buff := make([]byte, dataLen)
	n, err := hc.dec.DecodeDataFull(buff)
	if err != nil {
		hc.dec.Clean()
		return nil, err
	}

	p := &defs.Packet{}
	p.SetData(buff[:n])

	return p, nil
}

