/**
 * Created: 2020/3/26
 * @author: Jason
 */

package main

import (
	"flag"
	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/example/cluster/logic/app"
)

var srvName = flag.String("name", "logic", "server name")


func main() {
	flag.Parse()

	srv := app.NewLogicServer(*srvName, conf.GetConfPath())
	srv.Start()

	network.WaitExit()
}
