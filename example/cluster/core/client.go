/**
 * Created: 2020/4/13
 * @author: Jason
 */

package core

import (
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/defs"
)

type Client struct {
	*network.Session
}

func NewClient(conn defs.IConnection, sessionId string, serve defs.ServeObj, async ...bool) *Client {
	c := &Client{
		Session: network.NewSession(conn, sessionId, serve, async...),
	}
	return c
}
