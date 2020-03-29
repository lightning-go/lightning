/**
 * Created: 2020/3/27
 * @author: Jason
 */

package main

import (
	"flag"
	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/example/cluster/center/app"
)

var srvName = flag.String("name", "center", "server name")

func main() {
	flag.Parse()

	srv := app.NewCenterServer(*srvName, conf.GetConfPath())
	srv.Start()

	network.WaitExit()
}

