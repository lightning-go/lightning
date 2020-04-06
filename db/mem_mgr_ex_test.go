/**
 * Created: 2020/4/2
 * @author: Jason
 */

package db

import (
	"testing"
	"fmt"
	"github.com/json-iterator/go"
	"time"
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

	id := memUser.GetPKIncr()
	d := memUser.GetData(id)
	if d == nil {
		if id > 0 {
			u := User{
				Id:     id,
				Name:   "jason",
				Remark: 1,
				Test:   2323.2,
			}
			ok := memUser.AddData(u.Id, &u)
			if ok {
				memUser.SetDataIK("name", u.Name, u.Id)
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
			ok := memUser.UpdateData(u.Id, &u)
			if ok {
				memUser.SetDataIK("name", u.Name, u.Id)
				memUser.DelDataByIK("name", oldName)
			}
		}
	}

	time.Sleep(time.Second * 3)
	memUser.DelData(1)
	memUser.DelDataByIK("name", "jason22222")
}

func testMultiPK(rc *RedisClient) {
	memPlayer := NewMemMgrEx(rc, false, "test", "player", []string{"id", "name"})

	d := memPlayer.GetDataByMultiPK(map[string]interface{}{
		"id": 3,
		"name": "jason",
	})
	if d == nil {
		u := Player{
			Id:   3,
			Name: "jason",
			Info: "test...",
			Test: 2323.2,
		}
		memPlayer.AddDataByMultiPK(map[string]interface{}{
			"id": u.Id,
			"name": u.Name,
		}, &u)

	} else {
		u := Player{}
		err := jsoniter.Unmarshal(d, &u)
		if err != nil {
			fmt.Println("pk get", err)
		} else {
			fmt.Println("pk get", u.Id, u.Name, u.Info, u.Test)

			u.Test = 1231.1
			u.Info = "test...modify"
			memPlayer.UpdateDataByMultiPK(map[string]interface{}{
				"id": u.Id,
				"name": u.Name,
			}, &u)
		}
	}

	time.Sleep(time.Second * 3)
	memPlayer.DelDataByMultiPK(map[string]interface{}{
		"id": 3,
		"name": "jason",
	})
}

func Test1(t *testing.T) {
	dbMgr := NewMysqlMgr("test", "root", "123456", "127.0.0.1:3306")
	defer dbMgr.Close()

	rc := NewRedisClient("127.0.0.1:6379", 1, 0)
	defer rc.Close()

	testSinglePK(rc)

	time.Sleep(time.Second * 3)
}

func Test2(t *testing.T) {
	dbMgr := NewMysqlMgr("test", "root", "123456", "127.0.0.1:3306")
	defer dbMgr.Close()

	rc := NewRedisClient("127.0.0.1:6379", 1, 0)
	defer rc.Close()

	testMultiPK(rc)

	time.Sleep(time.Second * 3)
}