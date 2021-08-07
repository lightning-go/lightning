# lightning
a simple network server framework for golang, inspired by [muduo]

[muduo]:https://github.com/chenshuo/muduo


### [Example] 
[Example]:https://github.com/lightning-go/lightning/blob/master/example
* echo
  * echo server/client
* websocket
  * websocket server/client
* cluster
  * simple cluster, gate/logic/center
  * gate watches logic that registers to etcd
  * logic is connected to the center
* game server
  * distributed game server 
* pingpong
* ttcp



### Quick Start
* #### Echo Server

~~~golang
package main

import (
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/utils"
    //"github.com/lightning-go/lightning/example/echo/codec"
	"flag"
	"fmt"
)

var (
    port = flag.Int("p", 21000, "host port")
    maxConn = flag.Int("n", 3000, "max connection")
)

func main() {
    flag.Parse()

    host := fmt.Sprintf(":%v", *port)
    srv := network.NewTcpServer(host, "echo", *maxConn)

    srv.SetConnCallback(func(conn defs.IConnection) {
        logger.Tracef("%s server %s <- %s is %s %s",
            srv.Name(), conn.LocalAddr(), conn.RemoteAddr(),
            utils.IF(conn.IsClosed(), "down", "up"), " HAHAHA")
    })
    srv.SetMsgCallback(func(conn defs.IConnection, packet defs.IPacket) {
        logger.Tracef("onMsg: %s", packet.GetData())
        conn.WritePacket(packet)
    })	

    //srv.SetCodec(codec.NewHeadCodec())    //set your codec
    srv.Serve()

    network.WaitExit()
}
~~~

* #### Echo Client
~~~golang
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
    //"github.com/lightning-go/lightning/example/echo/codec"
)

var (
    host = flag.String("h", "127.0.0.1", "connect host")
    port = flag.Int("p", 21000, "connect port")
)

func main() {
    flag.Parse()	

    waitInput := make(chan bool, 1)
    addr := fmt.Sprintf("%v:%v", *host, *port)

    client := network.NewTcpClient("echo-client", addr)
    if client == nil {
        logger.Error("new client faield")
        return
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
    //client.SetCodec(codec.NewHeadCodec())  //set your codec
    client.Connect()	

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
~~~

### More
To be continued
