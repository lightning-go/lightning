/**
 * Created: 2019/4/22 0022
 * @author: Jason
 */

package module

import (
	"encoding/binary"
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

	hc.dec = NewDecoder(c, binary.BigEndian)
	hc.enc = NewEncode(c, binary.BigEndian)

	return true
}
func (hc *HeadCodec) Write(packet defs.IPacket) error {
	if hc.enc == nil {
		return ErrCodecWriteNil
	}

	//data len
	data := packet.GetData()
	dataLen := len(data)
	err := hc.enc.EncodeInt32(int32(dataLen))
	if err != nil {
		hc.enc.Clean()
		return err
	}

	//id len
	id := packet.GetId()
	idLen := len(id)
	err = hc.enc.EncodeInt32(int32(idLen))
	if err != nil {
		hc.enc.Clean()
		return err
	}
	if idLen > 0 {
		//id
		err = hc.enc.EncodeData([]byte(id))
		if err != nil {
			hc.enc.Clean()
			return err
		}
	}

	//session len
	sessionId := packet.GetSessionId()
	sIdLen := len(sessionId)
	err = hc.enc.EncodeInt32(int32(sIdLen))
	if err != nil {
		hc.enc.Clean()
		return err
	}
	if sIdLen > 0 {
		//sessionId
		err = hc.enc.EncodeData([]byte(sessionId))
		if err != nil {
			hc.enc.Clean()
			return err
		}
	}

	//sequence
	seq := packet.GetSequence()
	err = hc.enc.EncodeUInt64(seq)
	if err != nil {
		hc.enc.Clean()
		return err
	}

	//status
	err = hc.enc.EncodeInt32(int32(packet.GetStatus()))
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
		return nil, ErrCodecReadNil
	}

	//data len
	dataLen, err := hc.dec.DecodeInt32()
	if err != nil {
		hc.dec.Clean()
		return nil, err
	}

	//id len
	idLen, err := hc.dec.DecodeInt32()
	if err != nil {
		hc.dec.Clean()
		return nil, err
	}
	var id []byte
	if idLen > 0 {
		//id
		idData := make([]byte, idLen)
		n, err := hc.dec.DecodeDataFull(idData)
		if err != nil {
			hc.dec.Clean()
			return nil, err
		}
		id = idData[:n]
	}

	//sessionId len
	sIdLen, err := hc.dec.DecodeInt32()
	if err != nil {
		hc.dec.Clean()
		return nil, err
	}
	var sId []byte
	if sIdLen > 0 {
		//sessionId
		idData := make([]byte, sIdLen)
		n, err := hc.dec.DecodeDataFull(idData)
		if err != nil {
			hc.dec.Clean()
			return nil, err
		}
		sId = idData[:n]
	}

	//sequence
	seq, err := hc.dec.DecodeUInt64()
	if err != nil {
		hc.dec.Clean()
		return nil, err
	}

	//status
	status, err := hc.dec.DecodeInt32()
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
	p.SetId(string(id))
	p.SetSessionId(string(sId))
	p.SetSequence(uint64(seq))
	p.SetStatus(int(status))
	p.SetData(buff[:n])

	return p, nil
}

