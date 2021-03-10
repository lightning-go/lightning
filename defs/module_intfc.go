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
	SetId(interface{})
	GetId() string
	SetStatus(int)
	GetStatus() int
	GetSequence() uint64
	SetSequence(uint64)
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
	WriteAwait(IPacket) (IPacket, error)
	UpdateCodec(ICodec)
}
