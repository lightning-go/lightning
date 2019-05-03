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
	"lightning/common/util"
	"github.com/lightning-go/lightning/logger"
)

var (
	port = flag.Int("p", 12001, "listen port")
	path = flag.String("path", "/", "host path")
)

func main() {
	flag.Parse()

	addr := fmt.Sprintf(":%v", *port)
	srv := network.NewWSServer("websocket", addr, 5000, *path)
	if srv == nil {
		panic("alloc new server failed")
	}

	srv.SetConnCallback(func(conn defs.IConnection) {
		logger.Tracef("%v server %v <- %v is %v",
			srv.Name(), conn.LocalAddr(), conn.RemoteAddr(),
			util.IF(conn.IsClosed(), "down", "up"))
	})
	srv.SetMsgCallback(func(conn defs.IConnection, packet defs.IPacket) {
		logger.Tracef("onMsg: %s", packet.GetData())
		conn.WritePacket(packet)
	})
	srv.Serve()

	network.WaitExit()
}
