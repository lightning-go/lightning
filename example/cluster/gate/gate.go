/**
 * Created: 2020/3/25
 * @author: Jason
 */

package main

import (
	"flag"
	"github.com/lightning-go/lightning/example/cluster/gate/app"
	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/network"
)

var srvName = flag.String("name", "gate", "server name")

func main() {
	flag.Parse()

	srv := app.NewGateServer(*srvName, conf.GetConfPath())
	srv.Start()

	network.WaitExit()
}
