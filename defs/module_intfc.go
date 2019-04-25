/**
 * Created: 2019/4/20 0020
 * @author: Jason
 */

package defs

type IPacket interface {
	SetData([]byte)
	GetData() []byte
	SetSessionId(string)
	GetSessionId() string
	SetId(string)
	GetId() string
	SetStatus(int)
	GetStatus() int
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
