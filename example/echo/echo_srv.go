/**
 * Created: 2019/4/19 0019
 * @author: Jason
 */

package main

import (
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/utils"
	"github.com/lightning-go/lightning/module"
	"flag"
	"fmt"
	"github.com/lightning-go/lightning/conf"
)

var (
	port      = flag.Int("p", 21000, "host port")
	codecType = flag.Int("c", 1, "codec type: 1 stream, 2 head")
)

func main() {
	flag.Parse()

	var codec defs.ICodec = nil
	switch *codecType {
	case 1:
		codec = module.NewStreamCodec()
	case 2:
		codec = module.NewHeadCodec()
	}

	//
	host := fmt.Sprintf(":%v", *port)
	srv := network.NewTcpServer(host, "echo", conf.GetGlobalVal().MaxConnNum)

	srv.SetConnCallback(func(conn defs.IConnection) {
		logger.Tracef("%s server %s <- %s is %s %s",
			srv.Name(), conn.LocalAddr(), conn.RemoteAddr(),
			utils.IF(conn.IsClosed(), "down", "up"), " HAHAHA")
	})

	srv.SetMsgCallback(func(conn defs.IConnection, packet defs.IPacket) {
		logger.Tracef("onMsg: %s", packet.GetData())
		conn.WritePacket(packet)
	})

	srv.SetExitCallback(func() {
		logger.Tracef("%v exited", srv.Name())
	})

	srv.SetCodec(codec)
	srv.Serve()

	//
	//utils.WaitExit(srv.Stop)
	network.WaitExit()

}
