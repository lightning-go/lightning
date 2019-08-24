/**
 * Created: 2019/4/23 0023
 * @author: Jason
 */

package network

import (
	"github.com/lightning-go/lightning/defs"
	"net"
	"github.com/lightning-go/lightning/logger"
	"sync"
	"time"
)

type TcpClient struct {
	connector    *Connector
	conn         *Connection
	name         string
	codec        defs.ICodec
	ioModule     defs.IIOModule
	connCallback defs.ConnCallback
	msgCallback  defs.MsgCallback
	retry        bool
	connected    sync.WaitGroup
	timeout      time.Duration
	oneStep      bool
}

func NewTcpClient(name, addr string) *TcpClient {
	connector := NewConnector(addr)
	if connector == nil {
		return nil
	}
	client := &TcpClient{
		connector: connector,
		name:      name,
		retry:     true,
		oneStep:   false,
	}
	client.connector.SetConnCallback(client.connectionHandle)
	client.connector.SetCancelCallback(client.Cancel)
	return client
}

func (tcpClient *TcpClient) SetOneStep(val bool) {
	tcpClient.oneStep = val
}

func (tcpClient *TcpClient) GetConn() defs.IConnection {
	return tcpClient.conn
}

func (tcpClient *TcpClient) IsWorking() bool {
	return tcpClient.connector.IsWorking()
}

func (tcpClient *TcpClient) SetTimeout(val time.Duration) {
	tcpClient.timeout = val
}

func (tcpClient *TcpClient) SetCodec(codec defs.ICodec) {
	tcpClient.codec = codec
}

func (tcpClient *TcpClient) SetIOModule(ioModule defs.IIOModule) {
	tcpClient.ioModule = ioModule
}

func (tcpClient *TcpClient) SetConnCallback(cb defs.ConnCallback) {
	tcpClient.connCallback = cb
}

func (tcpClient *TcpClient) SetMsgCallback(cb defs.MsgCallback) {
	tcpClient.msgCallback = cb
}

func (tcpClient *TcpClient) Name() string {
	return tcpClient.name
}

func (tcpClient *TcpClient) SetRetry(v bool) {
	tcpClient.retry = v
}

func (tcpClient *TcpClient) connectionHandle(conn net.Conn) {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return
	}
	tcpConn.SetNoDelay(true)

	tcpClient.conn = NewConnection(conn)
	if tcpClient.conn == nil {
		return
	}
	tcpClient.conn.SetCodec(tcpClient.codec)
	tcpClient.conn.SetIOModule(tcpClient.ioModule)
	tcpClient.conn.SetCloseCallback(tcpClient.CloseConnection)
	tcpClient.conn.SetConnCallback(tcpClient.connCallback)
	tcpClient.conn.SetMsgCallback(tcpClient.msgCallback)

	if !tcpClient.oneStep {
		if !tcpClient.conn.Start() {
			return
		}
	}

	tcpClient.connected.Done()
}

func (tcpClient *TcpClient) Start() {
	if tcpClient.conn == nil || tcpClient.conn.IsClosed() {
		return
	}
	tcpClient.conn.Start()
}

func (tcpClient *TcpClient) Close() bool {
	tcpClient.retry = false
	return tcpClient.conn.Close()
}

func (tcpClient *TcpClient) Cancel() {
	tcpClient.connected.Done()
}

func (tcpClient *TcpClient) CloseConnection(conn defs.IConnection) {
	if conn != nil {
		logger.Tracef("close connection: %v", conn.GetId())
		conn.OnConnection()
	}
	if tcpClient.retry {
		tcpClient.connected.Add(1)
	}
	tcpClient.connector.Close(tcpClient.retry)
}

func (tcpClient *TcpClient) Connect() defs.IConnection {
	tcpClient.connected.Add(1)
	go tcpClient.connector.Start(tcpClient.timeout)
	tcpClient.connected.Wait()
	return tcpClient.conn
}

func (tcpClient *TcpClient) SendPacket(packet defs.IPacket) {
	tcpClient.conn.WritePacket(packet)
}

func (tcpClient *TcpClient) SendData(data []byte) {
	tcpClient.conn.WriteData(data)
}
