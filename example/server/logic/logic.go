/**
 * Created: 2019/4/24 0024
 * @author: Jason
 */

package main

import (
	"github.com/lightning-go/lightning/example/server/logic/app"
	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/network"
	"flag"
)

var (
	srvName = flag.String("name", "logic", "server name")
)

func main() {
	srv := app.NewGameServer(*srvName, conf.GetConfPath())
	srv.Start()

	network.WaitExit()
}
