/**
 * Created: 2020/4/2
 * @author: Jason
 */

package db

import (
	"testing"
	"fmt"
	"time"
	"github.com/jinzhu/gorm"
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

	u := User{}
	err := memUser.GetData(1, &u)
	if err == nil {
		fmt.Println("pk get", u.Id, u.Name, u.Remark, u.Test)
	}

	id := memUser.GetPKIncr()
	u = User{}
	err = memUser.GetData(id, &u)
	if err == gorm.ErrRecordNotFound {
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
		fmt.Println("pk get", u.Id, u.Name, u.Remark, u.Test)
	}

	u = User{}
	err = memUser.GetDataByIK("name", "jason", &u)
	if err == nil {
		fmt.Println("ik get", u.Id, u.Name, u.Remark, u.Test)

		oldName := u.Name
		u.Name = "jason22222"
		ok := memUser.UpdateData(u.Id, &u)
		if ok {
			memUser.SetDataIK("name", u.Name, u.Id)
			memUser.DelDataByIK("name", oldName)
		}
	}

	time.Sleep(time.Second * 3)
	memUser.DelData(1)
	memUser.DelDataByIK("name", "jason22222")
}

func testMultiPK(rc *RedisClient) {
	memPlayer := NewMemMgrEx(rc, false, "test", "player", []string{"id", "name"})

	p := Player{}
	err := memPlayer.GetDataByMultiPK(map[string]interface{}{
		"id": 3,
		"name": "jason",
	}, &p)
	if err == gorm.ErrRecordNotFound {
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
		fmt.Println("pk get", p.Id, p.Name, p.Info, p.Test)

		p.Test = 1231.1
		p.Info = "test...modify"
		memPlayer.UpdateDataByMultiPK(map[string]interface{}{
			"id": p.Id,
			"name": p.Name,
		}, &p)

	}

	err = memPlayer.GetDataByMultiPK(map[string]interface{}{
		"id": 3,
		"name": "jason",
	}, &p)
	if err == nil {
		fmt.Println("pk get", p.Id, p.Name, p.Info, p.Test)
	}

	time.Sleep(time.Second * 3)
	memPlayer.DelDataByMultiPK(map[string]interface{}{
		"id": 3,
		"name": "jason",
	})
}

func Test1(t *testing.T) {
	dbMgr := NewDBMgr(DB_type_mysql, "test", "root", "123456", "127.0.0.1:3306")
	defer dbMgr.Close()

	rc := NewRedisClient("127.0.0.1:6379", 1, 0)
	defer rc.Close()

	testSinglePK(rc)

	time.Sleep(time.Second * 3)
}

func Test2(t *testing.T) {
	dbMgr := NewDBMgr(DB_type_mysql, "test", "root", "123456", "127.0.0.1:3306")
	defer dbMgr.Close()

	rc := NewRedisClient("127.0.0.1:6379", 1, 0)
	defer rc.Close()

	testMultiPK(rc)

	time.Sleep(time.Second * 3)
}