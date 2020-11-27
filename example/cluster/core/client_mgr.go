/**
 * Created: 2020/4/15
 * @author: Jason
 */

package core

import (
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/network"
)

var defaultClientMgr = network.NewSessionMgr()

func GetClient(sessionId string) defs.ISession {
	return defaultClientMgr.GetSession(sessionId)
}

func DelClient(sessionId string) {
	defaultClientMgr.DelSession(sessionId)
}

func DelClientByConnId(connId string) []string {
	return defaultClientMgr.DelConnSession(connId)
}

func RangeClient(f func(string, defs.ISession) bool) {
	defaultClientMgr.RangeSession(f)
}

func GetClientCount() int64 {
	return defaultClientMgr.SessionCount()
}

//func CheckAddClient(conn defs.IConnection, sessionId string, serve defs.ServeObj, async ...bool) defs.ISession {
func CheckAddClient(conn defs.IConnection, sessionId string, serviceHandle network.ServiceHandle, async ...bool) defs.ISession {
	client := defaultClientMgr.GetSession(sessionId)
	if client == nil {
		//client = NewClient(conn, sessionId, serve, async...)
		client = NewClient(conn, sessionId, serviceHandle, async...)
		if client == nil {
			return nil
		}
		defaultClientMgr.AddSession(client)
	}
	return client
}
