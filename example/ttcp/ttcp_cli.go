/**
 * Created: 2019/4/23 0023
 * @author: Jason
 */

package main

import (
	"flag"
	"fmt"
	"github.com/lightning-go/lightning/defs"
	"time"
	"encoding/json"
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/logger"
	"github.com/json-iterator/go"
	"bytes"
	"sync"
)

var (
	host        = flag.String("h", "127.0.0.1", "connect host")
	port        = flag.Int("p", 10001, "connect port")
	count       = flag.Int("n", 1, "send count")
	clientCount = flag.Int("c", 1, "client number")

	msgData = []byte("hello world!!! hi Jason, it is a test!!!")

	mux             sync.RWMutex
	quit            sync.WaitGroup
	totalCostTime   float64
	totalAvge       float64
	totalBytes      int64
	totalCount      int64
	errorCount      int64
)

type TestMsg struct {
	Count int    `json:"count"`
	Data  []byte `json:"data"`
}

type TTcpClient struct {
	defs.IClient
	data       []byte
	totalCount int64
	errorCount int64
	byteRead   int64
	startTime  time.Time
	wait       chan bool
}

func NewTTcpClient(name string, addr string) *TTcpClient {
	c := &TTcpClient{
		IClient: network.NewTcpClient(name, addr),
		data:    make([]byte, 0),
		wait:    make(chan bool),
	}

	c.SetCodec(&module.HeadCodec{})
	c.SetConnCallback(c.onConn)
	c.SetMsgCallback(c.onMsg)

	msg := &TestMsg{0, msgData}
	c.data, _ = json.Marshal(msg)

	return c
}

func (tc *TTcpClient) onConn(conn defs.IConnection) {
}

func (tc *TTcpClient) onMsg(conn defs.IConnection, packet defs.IPacket) {
	data := packet.GetData()
	t := &TestMsg{}
	err := jsoniter.Unmarshal(data, t)
	if err != nil {
		fmt.Println("onMsg error:", err)
		return
	}

	if t.Count == *count {
		tc.wait <- true
		return
	}

	if !bytes.Equal(tc.data, data) {
		tc.errorCount++
		logger.Errorf("data not equal: %v", data)
	}

	tc.byteRead += int64(len(data))
	tc.totalCount++
}

func send(client *TTcpClient) {
	if client == nil {
		fmt.Println("client is nil")
		return
	}

	msg := &TestMsg{
		Count: *count,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error(err)
		return
	}

	client.SendData(data)
	client.startTime = time.Now()

	for i := 0; i < *count; i++ {
		msg := &TestMsg{
			Count: 0,
			Data:  msgData,
		}
		data, err = jsoniter.Marshal(msg)
		if err != nil {
			logger.Error(err)
			continue
		}
		client.SendData(data)
	}

}

func wait(client *TTcpClient) {
	if client == nil {
		logger.Error("client is nil")
		return
	}

	<-client.wait

	mux.Lock()

	costTime := time.Now().Sub(client.startTime).Seconds()
	avge := float64(*count) / costTime

	totalBytes += client.byteRead
	totalCount += client.totalCount
	errorCount += client.errorCount
	totalCostTime += costTime
	totalAvge += avge

	mux.Unlock()

	client.Close()
	quit.Done()
}

func main() {
	flag.Parse()

	logger.SetLevel(logger.INFO)

	fmt.Println("client:", *clientCount, ", send and recv times:", *count)
	addr := fmt.Sprintf("%v:%v", *host, *port)
	quit.Add(*clientCount)

	for i := 0; i < *clientCount; i++ {
		name := fmt.Sprintf("ttcp_cli_%v", i+1)
		client := NewTTcpClient(name, addr)
		client.SetCodec(&module.HeadCodec{})
		client.Connect()

		go send(client)
		go wait(client)
	}

	//
	fmt.Println("waiting...")
	quit.Wait()

	fmt.Printf("total avge cost time: %.6fs, %.2f/s, %.2fMiB/s\n",
		(totalCostTime)/float64(*clientCount),
		(totalAvge)/float64(*clientCount),
		float64(totalBytes)/(totalCostTime*1024*1024))
	fmt.Printf("total send and recv %d times, error %d times\n",
		totalCount, errorCount)

	fmt.Println("------------- exit -------------")
}
