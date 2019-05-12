/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package main

import (
	"testing"
	"logger/log"
	"github.com/lightning-go/lightning/defs"
	"fmt"
	"github.com/lightning-go/lightning/utils"
	"github.com/lightning-go/lightning/example/server/global"
	"github.com/json-iterator/go"
	"time"
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/logger"
)

func TestGame(t *testing.T)  {
	logger.SetLevel(logger.INFO)

	client := network.NewTcpClient("client", "127.0.0.1:22001")
	if client == nil {
		log.ERROR("new client faield")
		return
	}

	client.SetConnCallback(func(conn defs.IConnection) {
		isClose := conn.IsClosed()
		fmt.Println(client.Name(), conn.LocalAddr(), "->", conn.RemoteAddr(), "is",
			utils.IF(isClose, "down", "up"))
	})

	client.SetMsgCallback(func(conn defs.IConnection, packet defs.IPacket) {
		fmt.Println("onMsg - status:", packet.GetStatus(),
			", sessionId:", packet.GetSessionId(),
			", data:", string(packet.GetData()))
	})

	client.SetCodec(&module.HeadCodec{})
	client.Connect()

	req := &global.MsgTestReq{
		N: 23435,
	}
	data, err := jsoniter.Marshal(req)
	if err != nil {
		log.ERROR(err)
		return
	}

	p := &defs.Packet{}
	p.SetId("Test")
	p.SetData(data)
	client.SendPacket(p)

	time.Sleep(time.Second * 13)
	client.Close()
}
