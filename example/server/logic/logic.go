/**
 * Created: 2019/4/24 0024
 * @author: Jason
 */

package main

import (
	"github.com/lightning-go/lightning/example/server/logic/app"
	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/network"
)

func main() {
	srv := app.NewGameServer("logic", conf.GetConfPath())
	srv.Start()

	network.WaitExit()
}
