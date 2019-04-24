/**
 * Created: 2019/4/23 0023
 * @author: Jason
 */

package main

import (
	"flag"
	"github.com/lightning-go/lightning/network"
	"fmt"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"lightning/common/util"
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
		logger.Trace("%v %v -> %v is %v\n", client.Name(),
			conn.LocalAddr(), conn.RemoteAddr(),
			util.IF(conn.IsClosed(), "down", "up"))

		data := utils.NowTimeFormat()
		client.SendData([]byte(data))
	})
	
	client.SetMsgCallback(func(conn defs.IConnection, packet defs.IPacket) {
		logger.Trace("onMsg: %s", packet.GetData())

		time.Sleep(time.Second * 1)
		data := util.NowTimeFormat()
		client.SendData([]byte(data))
	})

	client.Connect()

	utils.WaitExit()
}
