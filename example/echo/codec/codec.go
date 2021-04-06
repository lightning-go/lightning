package codec

import (
	"encoding/binary"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/module"
)

type HeadCodec struct {
	dec *module.Decoder
	enc *module.Encoder
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

	hc.dec = module.NewDecoder(c, binary.BigEndian)
	hc.enc = module.NewEncode(c, binary.BigEndian)

	return true
}
func (hc *HeadCodec) Write(packet defs.IPacket) error {
	if hc.enc == nil {
		return module.ErrCodecWriteNil
	}

	//data len
	data := packet.GetData()
	dataLen := len(data)
	err := hc.enc.EncodeInt32(int32(dataLen))
	if err != nil {
		hc.enc.Clean()
		return err
	}

	//data
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
		return nil, module.ErrCodecReadNil
	}

	//data len
	dataLen, err := hc.dec.DecodeInt32()
	if err != nil {
		hc.dec.Clean()
		return nil, err
	}

	//data
	buff := make([]byte, dataLen)
	n, err := hc.dec.DecodeDataFull(buff)
	if err != nil {
		hc.dec.Clean()
		return nil, err
	}

	p := &defs.Packet{}
	p.SetId("")
	p.SetSessionId("")
	p.SetSequence(0)
	p.SetStatus(0)
	p.SetData(buff[:n])

	return p, nil
}
