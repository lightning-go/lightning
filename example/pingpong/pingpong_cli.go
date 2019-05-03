/**
 * Created: 2019/4/24 0024
 * @author: Jason
 */

package main

import (
	"flag"
	"github.com/lightning-go/lightning/utils"
	"time"
	"fmt"
	"github.com/lightning-go/lightning/defs"
	"sync/atomic"
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/logger"
)

var (
	host      = flag.String("h", "127.0.0.1", "connect host")
	port      = flag.Int("p", 13001, "connect port")
	blockSize = flag.Int64("bs", 16, "send block size")
	clientNum = flag.Int("c", 1, "client number")
	timeout   = flag.Int("t", 60, "connect timeout")
	codecType = flag.Int("ct", 0, "codec type, 0 stream, 1 head")
)

type Session struct {
	defs.IClient
	client       *PingPongClient
	bytesRead    int64
	bytesWritten int64
	messagesRead int64
}

func NewSession(name, addr string, client *PingPongClient) *Session {
	if client == nil {
		panic("client is nil")
	}

	s := &Session{
		IClient: network.NewTcpClient(name, addr),
		client:  client,
	}

	s.SetCodec(client.codec)
	s.SetConnCallback(s.onConn)
	s.SetMsgCallback(s.onMsg)
	return s
}

func (s *Session) BytesRead() int64 {
	return s.bytesRead
}

func (s *Session) MessagesRead() int64 {
	return s.messagesRead
}

func (s *Session) onConn(conn defs.IConnection) {
	if conn.IsClosed() {
		s.client.onDisconn(conn)
	} else {
		conn.WriteData(s.client.MsgData())
		s.client.onConn()
	}
}

func (s *Session) onMsg(conn defs.IConnection, packet defs.IPacket) {
	data := packet.GetData()
	dataLen := int64(len(data))
	s.bytesRead += dataLen
	s.bytesWritten += dataLen
	s.messagesRead++
	conn.WriteData(data)
}

//////////////////////////////////////////////////////////////////////////////

type PingPongClient struct {
	name      string
	addr      string
	connNum   int32
	clientNum int
	blockSize int64
	timeout   int
	codec     defs.ICodec
	msg       []byte
	quit      chan bool
	sessions  []*Session
	endTimer  *utils.Timer
}

func NewPingPongClient(name, addr string, codecType, clientNum, timeout int, blockSize int64) *PingPongClient {
	ppc := &PingPongClient{
		name:      name,
		addr:      addr,
		clientNum: clientNum,
		blockSize: blockSize,
		timeout:   timeout,
		quit:      make(chan bool),
	}

	ppc.codec = &module.StreamCodec{}
	if codecType > 0 {
		ppc.codec = &module.HeadCodec{}
	}

	return ppc
}

func (p *PingPongClient) Start() {
	var i int64 = 0
	for ; i < p.blockSize; i++ {
		p.msg = append(p.msg, 'a')
	}

	for i := 0; i < p.clientNum; i++ {
		name := fmt.Sprintf("%v_%05d", p.name, i)
		session := NewSession(name, p.addr, p)
		if session == nil {
			continue
		}
		session.Connect()
		p.sessions = append(p.sessions, session)
	}

}

func (p *PingPongClient) MsgData() []byte {
	return p.msg
}

func (p *PingPongClient) timeoutHandle() {
	for _, session := range p.sessions {
		if session == nil {
			continue
		}
		session.Close()
	}
}

func (p *PingPongClient) onConn() {
	val := atomic.AddInt32(&p.connNum, 1)
	if val == int32(p.clientNum) {
		fmt.Println("all client connected ", p.addr)
		p.endTimer = utils.NewTimer(time.Duration(p.timeout)*time.Second, p.timeoutHandle)
		p.endTimer.Start(true)
	}
}

func (p *PingPongClient) onDisconn(conn defs.IConnection) {
	val := atomic.AddInt32(&p.connNum, -1)
	if val > 0 {
		return
	}
	fmt.Println("all client disconnected")

	var totalBytesRead int64
	var totalMsgRead int64

	for _, session := range p.sessions {
		if session == nil {
			continue
		}
		totalBytesRead += session.BytesRead()
		totalMsgRead += session.MessagesRead()
	}

	fmt.Println(totalMsgRead, " total messages read")
	fmt.Println(float64(totalBytesRead)/float64(totalMsgRead),
		" average message size")
	fmt.Println(float64(totalBytesRead)/(1024*1024),
		" MiB total bytes read")
	fmt.Println(float64(totalMsgRead)/float64(p.timeout),
		"/s message")
	fmt.Println(float64(totalBytesRead)/float64(p.timeout*1024*1024),
		" MiB/s throughput")

	p.quit <- true
}

func main() {
	flag.Parse()

	logger.SetLevel(logger.INFO)

	addr := fmt.Sprintf("%v:%v", *host, *port)
	client := NewPingPongClient("pingpong", addr, *codecType, *clientNum, *timeout, *blockSize)
	client.Start()

	<-client.quit
	fmt.Println("---------------- exit ----------------")
}
