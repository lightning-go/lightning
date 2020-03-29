/**
 * Created: 2019/4/23 0023
 * @author: Jason
 */

package main

import (
	"flag"
	"fmt"
	"bufio"
	"os"
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/utils"
	"github.com/lightning-go/lightning/module"
)

var (
	host      = flag.String("h", "127.0.0.1", "connect host")
	port      = flag.Int("p", 21000, "connect port")
)


func createClient(addr string, codec defs.ICodec, waitInput chan bool) *network.TcpClient {
	client := network.NewTcpClient("echo-client", addr)
	if client == nil {
		logger.Error("new client faield")
		return nil
	}

	client.SetConnCallback(func(conn defs.IConnection) {
		isClose := conn.IsClosed()
		fmt.Printf("%v %v -> %v is %v \n>> ", client.Name(),
			conn.LocalAddr(), conn.RemoteAddr(),
			utils.IF(isClose, "down", "up"))
		if !isClose {
			waitInput <- true
		}
	})

	client.SetCodec(codec)
	client.Connect()
	return client
}


func main() {
	flag.Parse()

	codec := module.NewHeadCodec()
	waitInput := make(chan bool, 1)
	addr := fmt.Sprintf("%v:%v", *host, *port)

	client := createClient(addr, codec, waitInput)
	if client == nil {
		fmt.Println("create client failed")
		return
	}

	//
	<-waitInput
	fmt.Println("Please input:")
	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">> ")
		rawLine, _, _ := r.ReadLine()
		line := string(rawLine)
		if len(line) == 0 {
			continue
		}
		if line == "q" || line == "Q" {
			client.Close()
			break
		}

		ack, err := client.SendDataAwait(rawLine)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("await ack %v, %s\n", ack.GetSequence(), ack.GetData())
	}

	fmt.Println("--------- exit ---------")
}
