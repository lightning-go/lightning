/**
 * Created: 2019/4/20 0020
 * @author: Jason
 */

package defs

type Packet struct {
	sessionId string
	id        uint32
	data      []byte
}

func (p *Packet) GetSessionId() string {
	return p.sessionId
}

func (p *Packet) SetSessionId(sessionId string) {
	p.sessionId = sessionId
}

func (p *Packet) GetId() uint32 {
	return p.id
}

func (p *Packet) SetId(id uint32) {
	p.id = id
}

func (p *Packet) GetData() []byte {
	return p.data
}

func (p *Packet) SetData(data []byte) {
	p.data = data
}
