/**
 * Created: 2019/4/23 0023
 * @author: Jason
 */

package main

import (
	"flag"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/logger"
	"github.com/json-iterator/go"
	"fmt"
)

var (
	port       = flag.Int("p", 10001, "Listen port")
	connLimit  = flag.Int("n", 50000, "Client connect limit")
	counterKey = "counter"
)

type TestMsg struct {
	Count int    `json:"count"`
	Data  []byte `json:"data"`
}

type TTcpServer struct {
	defs.IServer
	count int
}

func NewTTcpServer(name, addr string, maxConn int) *TTcpServer {
	s := &TTcpServer{
		IServer: network.NewTcpServer(addr, name, maxConn),
	}
	s.SetCodec(&module.HeadCodec{})
	s.SetConnCallback(s.onConn)
	s.SetMsgCallback(s.onMsg)
	return s
}

func (ts *TTcpServer) onConn(conn defs.IConnection) {
}

func (ts *TTcpServer) onMsg(conn defs.IConnection, packet defs.IPacket) {
	data := packet.GetData()

	t := &TestMsg{}
	err := jsoniter.Unmarshal(data, t)
	if err != nil {
		logger.Error(err)
		conn.Close()
		return
	}

	tmpCount := t.Count
	if tmpCount > 0 {
		ts.count = tmpCount
		conn.SetContext(counterKey, int(0))
		return
	}

	conn.WritePacket(packet)

	iCounter := conn.GetContext(counterKey)
	if iCounter == nil {
		conn.Close()
		return
	}
	counter, ok := iCounter.(int)
	if !ok {
		conn.Close()
		return
	}

	counter++
	conn.SetContext(counterKey, counter)
	if counter < ts.count {
		return
	}

	d := &TestMsg{
		Count: ts.count,
	}
	data, err = jsoniter.Marshal(d)
	if err != nil {
		logger.Error(err)
		return
	}

	p := &defs.Packet{}
	p.SetData(data)
	conn.WritePacket(p)

}

func main() {
	flag.Parse()

	logger.SetLevel(logger.INFO)

	addr := fmt.Sprintf(":%v", *port)
	srv := NewTTcpServer("ttcp", addr, *connLimit)
	srv.Serve()

	network.WaitExit()
}
