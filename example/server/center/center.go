/**
 * @author: Jason
 * Created: 19-5-12
 */

package main

import (
	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/example/server/center/app"
	"flag"
)

var (
	srvName = flag.String("name", "center", "server name")
)

func main() {
	srv := app.NewCenterServer(*srvName, conf.GetConfPath())
	srv.Start()

	network.WaitExit()
}



