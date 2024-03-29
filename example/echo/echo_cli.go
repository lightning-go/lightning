/**
 * Created: 2019/4/23 0023
 * @author: Jason
 */

package main

import (
	"flag"
	"fmt"
	"bufio"
	"github.com/lightning-go/lightning/example/echo/codec"
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
	codecType = flag.Int("c", 1, "codec type: 1 stream, 2 head")
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

	client.SetMsgCallback(func(conn defs.IConnection, packet defs.IPacket) {
		fmt.Printf("onMsg: %s (%d)\n>> ", packet.GetData(), packet.GetSequence())
	})

	client.SetCodec(codec)
	client.Connect()
	return client
}


func main() {
	flag.Parse()

	var c defs.ICodec = nil
	switch *codecType {
	case 1:
		c = module.NewStreamCodec()
	case 2:
		c = codec.NewHeadCodec()
	}

	waitInput := make(chan bool, 1)
	addr := fmt.Sprintf("%v:%v", *host, *port)

	client := createClient(addr, c, waitInput)
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

		client.SendData(rawLine)
	}

	fmt.Println("--------- exit ---------")
}
