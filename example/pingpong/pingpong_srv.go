/**
 * Created: 2019/4/24 0024
 * @author: Jason
 */

package main

import (
	"flag"
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/logger"
	"fmt"
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/defs"
)

var (
	port      = flag.Int("p", 13001, "listen port")
	connLimit = flag.Int("n", 50000, "client connnect limit")
	codecType = flag.Int("ct", 0, "codec type, 0 stream, 1 head")
)

func main() {
	flag.Parse()

	logger.SetLevel(logger.INFO)

	addr := fmt.Sprintf(":%v", *port)
	srv := network.NewTcpServer(addr, "pingpong", *connLimit)
	if srv == nil {
		panic("new server failed")
	}

	var codec defs.ICodec = &module.StreamCodec{}
	if *codecType > 0 {
		codec = &module.HeadCodec{}
	}

	srv.SetMsgCallback(func(conn defs.IConnection, packet defs.IPacket) {
		conn.WritePacket(packet)
	})
	srv.SetCodec(codec)
	srv.Serve()

	network.WaitExit()
}
