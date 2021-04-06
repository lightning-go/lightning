/**
 * Created: 2019/4/23 0023
 * @author: Jason
 */

package main

import (
	"flag"
	"fmt"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/utils"
	"time"
)

var (
	host = flag.String("h", "127.0.0.1", "connect host")
	port = flag.Int("p", 12001, "connect port")
	path = flag.String("path", "/", "host path")
)

func main() {
	flag.Parse()

	addr := fmt.Sprintf("%v:%v", *host, *port)

	client := network.NewWSClient("wsclient", addr, *path)
	if client == nil {
		panic("alloc new client failed")
	}

	client.SetConnCallback(func(conn defs.IConnection) {
		logger.Tracef("%v %v -> %v is %v\n", client.Name(),
			conn.LocalAddr(), conn.RemoteAddr(),
			utils.IF(conn.IsClosed(), "down", "up"))

		data := utils.NowTimeFormat()
		client.SendData([]byte(data))
	})

	client.SetMsgCallback(func(conn defs.IConnection, packet defs.IPacket) {
		logger.Trace("", logger.Fields{"onMsg": string(packet.GetData())})

		time.Sleep(time.Second * 1)
		data := utils.NowTimeFormat()
		client.SendData([]byte(data))
	})

	client.Connect()

	utils.WaitExit()
}
