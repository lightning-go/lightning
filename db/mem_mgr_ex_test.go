/**
 * Created: 2020/4/2
 * @author: Jason
 */

package db

import (
	"testing"
	"fmt"
	"github.com/json-iterator/go"
)

type User struct {
	Id     int64   `json:"id"`
	Name   string  `json:"name"`
	Remark int     `json:"remark"`
	Test   float32 `json:"test"`
}

type Player struct {
	Id   int64   `json:"id"`
	Name string  `json:"name"`
	Info string  `json:"info"`
	Test float32 `json:"test"`
}

func testSinglePK(rc *RedisClient) {
	memUser := NewMemMgrEx(rc, true, "test", "user", []string{"id"})
	defer memUser.Close()

	d := memUser.GetData(46)
	if d == nil {
		id := memUser.GetPKIncr()
		if id > 0 {
			u := User{
				Id:     id,
				Name:   "jason",
				Remark: 1,
				Test:   2323.2,
			}
			ok := memUser.SetData(u.Id, &u)
			if ok {
				memUser.SetDataByIK("name", u.Name, u.Id)
			}
		}

	} else {
		u := User{}
		err := jsoniter.Unmarshal(d, &u)
		if err != nil {
			fmt.Println("pk get", err)
		} else {
			fmt.Println("pk get", u.Id, u.Name, u.Remark, u.Test)
		}
	}

	d = memUser.GetDataByIK("name", "jason")
	if d != nil {
		u := User{}
		err := jsoniter.Unmarshal(d, &u)
		if err != nil {
			fmt.Println("ik get", err)
		} else {
			fmt.Println("ik get", u.Id, u.Name, u.Remark, u.Test)

			oldName := u.Name
			u.Name = "jason22222"
			ok := memUser.SetData(u.Id, &u)
			if ok {
				memUser.SetDataByIK("name", u.Name, u.Id)
				memUser.DelDataByIK("name", oldName)
			}
		}
	}

	memUser.DelData(46)
	memUser.DelDataByIK("name", "jason22222")
}

func testMultiPK(rc *RedisClient) {
	memPlayer := NewMemMgrEx(rc, false, "test", "player", []string{"id", "name"})
	defer memPlayer.Close()

	d := memPlayer.GetDataByMultiPK(3, "jason")
	if d == nil {
		u := Player{
			Id:   3,
			Name: "jason",
			Info: "test...",
			Test: 2323.2,
		}
		memPlayer.SetDataByMultiPK(&u, u.Id, u.Name)

	} else {
		u := Player{}
		err := jsoniter.Unmarshal(d, &u)
		if err != nil {
			fmt.Println("pk get", err)
		} else {
			fmt.Println("pk get", u.Id, u.Name, u.Info, u.Test)

			u.Test = 1231.1
			u.Info = "test...modify"
			memPlayer.SetDataByMultiPK(&u, u.Id, u.Name)
		}
	}

	memPlayer.DelDataByMultiPK(3, "jason")
}

func Test1(t *testing.T) {
	dbMgr := NewMysqlMgr("test", "root", "123456", "127.0.0.1:3306")
	defer dbMgr.Close()

	rc := NewRedisClient("127.0.0.1:6379", 1, 0)
	testSinglePK(rc)
	testMultiPK(rc)
}
