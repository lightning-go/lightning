/**
 * Created: 2019/4/23 0023
 * @author: Jason
 */

package module

import (
	"time"
	"github.com/gorilla/websocket"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/conf"
)

type WSCodec struct {
	conn *websocket.Conn
}

func NewWSCodec() *WSCodec {
	return &WSCodec{}
}

func (wsCodec *WSCodec) Init(conn defs.IConnection) bool {
	if conn == nil {
		return false
	}
	iWSConn, ok := conn.(defs.IWSConnection)
	if !ok {
		return false
	}
	c := iWSConn.GetConn()
	if c == nil {
		return false
	}
	wsCodec.conn = c
	return true
}
func (wsCodec *WSCodec) Write(packet defs.IPacket) error {
	err := wsCodec.conn.SetWriteDeadline(time.Now().Add(conf.GetGlobalVal().WriteWait))
	if err != nil {
		return err
	}

	err = wsCodec.conn.WriteMessage(websocket.BinaryMessage, packet.GetData())
	if err != nil {
		err = wsCodec.conn.WriteMessage(websocket.CloseMessage, []byte{})
	}
	return nil
}

func (wsCodec *WSCodec) Read() (defs.IPacket, error) {
	_, data, err := wsCodec.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	p := &defs.Packet{}
	p.SetId(0)
	p.SetData(data)

	return p, nil
}


