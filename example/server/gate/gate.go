/**
 * Created: 2019/4/24 0024
 * @author: Jason
 */

package main

import (
	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/example/server/gate/app"
)

func main() {
	srv := app.NewGateServer("gate", conf.GetConfPath())
	srv.Start()

	network.WaitExit()
}

