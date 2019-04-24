/**
 * Created: 2019/4/20 0020
 * @author: Jason
 */

package defs

type IPacket interface {
	SetData(data []byte)
	GetData() []byte
	SetSessionId(sessionId string)
	GetSessionId() string
	SetId(id uint32)
	GetId() uint32
}

type ICodec interface {
	Init(IConnection) bool
	Write(IPacket) error
	Read() (IPacket, error)
}

type IIOModule interface {
	Codec(ICodec) bool
	Close()
	OnConnectionLost()
	Write(IPacket)
}
