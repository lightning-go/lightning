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

//func NewClient(conn defs.IConnection, sessionId string, serve defs.ServeObj, async ...bool) *Client {
func NewClient(conn defs.IConnection, sessionId string, serviceHandle network.ServiceHandle, async ...bool) *Client {
	c := &Client{
		//Session: network.NewSession(conn, sessionId, serve, async...),
		Session: network.NewSession(conn, sessionId, serviceHandle, async...),
	}
	return c
}
