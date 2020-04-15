/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package main

import (
	"testing"
	"github.com/lightning-go/lightning/defs"
	"fmt"
	"github.com/lightning-go/lightning/utils"
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/module"
	"time"
	"github.com/lightning-go/lightning/example/cluster/msg"
	"github.com/lightning-go/lightning/example/cluster/common"
)

func TestGame(t *testing.T)  {
	waitInput := make(chan bool, 1)

	client := network.NewTcpClient("client", "127.0.0.1:22001")
	if client == nil {
		fmt.Println("new client faield")
		return
	}
	defer client.Close()

	client.SetConnCallback(func(conn defs.IConnection) {
		closed := conn.IsClosed()
		fmt.Println(client.Name(), conn.LocalAddr(), "->", conn.RemoteAddr(), "is",
			utils.IF(closed, "down", "up"))
		if !closed {
			waitInput <- true
		}
	})

	client.SetMsgCallback(func(connection defs.IConnection, packet defs.IPacket) {
		fmt.Println("onMsg", packet.GetSequence(), packet.GetId(), string(packet.GetData()))
	})

	client.SetCodec(&module.HeadCodec{})
	client.Connect()

	<- waitInput

	p := &defs.Packet{}
	p.SetId("Ping")
	p.SetData([]byte("{}"))
	client.SendPacket(p)

	tr := &msg.MsgTestReq{N: 2}
	d := common.MarshalDataEx(tr)
	p1 := &defs.Packet{}
	p1.SetId("Test")
	p1.SetData(d)
	client.SendPacket(p1)

	tr2 := &msg.MsgTest2Req{N: 3}
	d = common.MarshalDataEx(tr2)
	p2 := &defs.Packet{}
	p2.SetId("Test2")
	p2.SetData(d)
	client.SendPacket(p2)

	tr3 := &msg.MsgTestCenterReq{N: 3}
	d = common.MarshalDataEx(tr3)
	p3 := &defs.Packet{}
	p3.SetId("Test3")
	p3.SetData(d)
	client.SendPacket(p3)


	time.Sleep(time.Second * 10)
}
