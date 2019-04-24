/**
 * Created: 2019/4/23 0023
 * @author: Jason
 */

package network

import (
	"github.com/lightning-go/lightning/defs"
	"net"
	"time"
	"github.com/lightning-go/lightning/logger"
)

type Connector struct {
	addr         string
	close        chan bool
	working      bool
	connCallback defs.ClientConnCallback
}

func NewConnector(addr string) *Connector {
	return &Connector{
		addr:  addr,
		close: make(chan bool),
	}
}

func (c *Connector) SetConnCallback(cb defs.ClientConnCallback) {
	c.connCallback = cb
}

func (c *Connector) connectionHandle(conn net.Conn) {
	if c.connCallback != nil {
		c.connCallback(conn)
	}
}

func (c *Connector) Close(v bool) {
	c.close <- v
}

func (c *Connector) Start() {
	if c.working {
		return
	}
	c.connect(c.addr)
}

func (c *Connector) connect(addr string) {
	c.working = true
	var tmpDelay time.Duration
	maxDelay := 3 * time.Second

	for {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			if tmpDelay == 0 {
				tmpDelay = time.Second
			} else {
				tmpDelay += time.Second
			}
			if tmpDelay > maxDelay {
				tmpDelay = maxDelay
			}
			logger.Warn("connecting to %v error, retrying in %v second", addr, tmpDelay.Seconds())
			time.Sleep(tmpDelay)
			continue
		}

		go c.connectionHandle(conn)

		retry := <-c.close
		if !retry {
			break
		}

		tmpDelay = 0
		logger.Warn("reconnecting to %v", addr)
	}

	c.working = false
}
