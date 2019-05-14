/**
 * Created: 2019/4/23 0023
 * @author: Jason
 */

package network

import (
	"github.com/gorilla/websocket"
	"github.com/lightning-go/lightning/defs"
	"sync"
	"github.com/lightning-go/lightning/logger"
	"net/url"
	"time"
)

type WSClient struct {
	connector    *websocket.Dialer
	conn         *WSConnection
	name         string
	addr         string
	path         string
	msgType      int
	codec        defs.ICodec
	ioModule     defs.IIOModule
	connCallback defs.ConnCallback
	msgCallback  defs.MsgCallback
	retry        bool
	connected    sync.WaitGroup
	close        chan bool
}

func NewWSClient(name, addr string, path ...string) *WSClient {
	client := &WSClient{
		connector: &websocket.Dialer{},
		conn:      nil,
		name:      name,
		addr:      addr,
		retry:     true,
		msgType:   websocket.TextMessage,
		close:     make(chan bool),
	}
	if len(path) > 0 {
		client.path = path[0]
	}
	return client
}

func (wsclient *WSClient) SetCodec(codec defs.ICodec) {
	wsclient.codec = codec
}

func (wsclient *WSClient) SetIOModule(ioModule defs.IIOModule) {
	wsclient.ioModule = ioModule
}

func (wsclient *WSClient) SetConnCallback(cb defs.ConnCallback) {
	wsclient.connCallback = cb
}

func (wsclient *WSClient) SetMsgCallback(cb defs.MsgCallback) {
	wsclient.msgCallback = cb
}

func (wsclient *WSClient) Name() string {
	return wsclient.name
}

func (wsclient *WSClient) SetRetry(v bool) {
	wsclient.retry = v
}

func (wsclient *WSClient) connectionHandle(conn *websocket.Conn) {
	if conn == nil {
		return
	}
	wsclient.conn = NewWSConnection(conn)
	if wsclient.conn == nil {
		return
	}
	wsclient.conn.SetMsgType(wsclient.msgType)
	wsclient.conn.SetCodec(wsclient.codec)
	wsclient.conn.SetIOModule(wsclient.ioModule)
	wsclient.conn.SetCloseCallback(wsclient.CloseConnection)
	wsclient.conn.SetConnCallback(wsclient.connCallback)
	wsclient.conn.SetMsgCallback(wsclient.msgCallback)

	if !wsclient.conn.Start() {
		return
	}

	wsclient.connected.Done()
}

func (wsclient *WSClient) Close() bool {
	wsclient.retry = false
	return wsclient.conn.Close()
}

func (wsclient *WSClient) CloseConnection(conn defs.IConnection) {
	if conn != nil {
		logger.Tracef("CloseConnection: %v", conn.GetId())
		conn.OnConnection()
	}
	if wsclient.retry {
		wsclient.connected.Add(1)
	}
	wsclient.close <- wsclient.retry
}

func (wsclient *WSClient) Connect() defs.IConnection {
	wsclient.connected.Add(1)
	go wsclient.connect()
	wsclient.connected.Wait()
	return wsclient.conn
}

func (wsclient *WSClient) connect() {
	u := url.URL{
		Scheme: "ws",
		Host:   wsclient.addr,
		Path:   wsclient.path,
	}
	addr := u.String()

	var tmpDelay time.Duration
	maxDelay := 3 * time.Second

	for {
		conn, _, err := wsclient.connector.Dial(addr, nil)
		if err != nil {
			if tmpDelay == 0 {
				tmpDelay = time.Second
			} else {
				tmpDelay += time.Second
			}
			if tmpDelay > maxDelay {
				tmpDelay = maxDelay
			}
			logger.Warnf("connecting to %v error, retrying in %v second", addr, tmpDelay.Seconds())
			time.Sleep(tmpDelay)
			continue
		}

		go wsclient.connectionHandle(conn)

		retry := <-wsclient.close
		if !retry {
			break
		}

		tmpDelay = 0
		logger.Warnf("reconnecting to %v", addr)
	}

	logger.Warn("connection disconnected")
}

func (wsclient *WSClient) SendPacket(packet defs.IPacket) {
	wsclient.conn.WritePacket(packet)
}

func (wsclient *WSClient) SendData(data []byte) {
	wsclient.conn.WriteData(data)
}
