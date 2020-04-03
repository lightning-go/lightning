/**
 * Created: 2019/6/15
 * @author: Jason
 */

package main

import (
	"github.com/lightning-go/lightning/db"
	"fmt"
	"strconv"
	"github.com/jinzhu/gorm"
	"database/sql"
)

type User struct {
	Id     int     `gorm:"column:id"`
	Name   string  `gorm:"column:name"`
	Remark int     `gorm:"column:remark"`
	Test   float32 `gorm:"column:test"`
}

func singleKeyTest(rc *db.RedisClient) {
	if rc == nil {
		return
	}

	mem := db.NewMemMgr(rc, false, "test", "user", []string{"id"}, "name")
	mem.LoadDB()

	//查
	memMode := mem.QueryCheckDBByPK(strconv.Itoa(2))
	if memMode != nil {
		id := memMode.Data["id"]
		name := memMode.Data["name"]
		remark := memMode.Data["remark"]
		test := memMode.Data["test"]
		fmt.Println("query:", id, name, remark, test)

		//改
		m := mem.UpdateByPk("2", "name", "jim")
		id = m.Data["id"]
		name = m.Data["name"]
		remark = m.Data["remark"]
		test = m.Data["test"]
		fmt.Println("query:", id, name, remark, test)
		mem.SyncMemMode(m)

		m = mem.UpdateByPk("2", "test", 2.35)
		id = m.Data["id"]
		name = m.Data["name"]
		remark = m.Data["remark"]
		test = m.Data["test"]
		fmt.Println("query:", id, name, remark, test)
		mem.SyncMemMode(m)
	}

	memMode = mem.QueryCheckDBByIK("name", "jerry")
	if memMode != nil {
		id := memMode.Data["id"]
		name := memMode.Data["name"]
		remark := memMode.Data["remark"]
		fmt.Println("query ik[name]:", id, name, remark)

		mem.UpdateByPk("1", "name", "xxx")
	}

	//增
	memMode = db.NewMemMode()
	memMode.Data["id"] = 3
	memMode.Data["name"] = "Jason"
	memMode.Data["remark"] = 1
	memMode.Data["test"] = 5.01
	mem.AddData(memMode, false)
	mem.SyncMemMode(memMode)

	memMode = db.NewMemMode()
	memMode.Data["id"] = 4
	memMode.Data["name"] = "Toney"
	memMode.Data["remark"] = 1
	memMode.Data["test"] = 1.5
	mem.AddData(memMode, false)

	memMode = mem.QueryCheckDBByPK(strconv.Itoa(3))
	if memMode != nil {
		id := memMode.Data["id"]
		name := memMode.Data["name"]
		remark := memMode.Data["remark"]
		test := memMode.Data["test"]
		fmt.Println("query:", id, name, remark, test)
	}
	memMode = mem.QueryCheckDBByPK(strconv.Itoa(4))
	if memMode != nil {
		id := memMode.Data["id"]
		name := memMode.Data["name"]
		remark := memMode.Data["remark"]
		test := memMode.Data["test"]
		fmt.Println("query:", id, name, remark, test)
	}

	//删
	mem.DelDataCheckDBByPK("3")
	memMode = mem.QueryCheckDBByPK(strconv.Itoa(3))
	if memMode != nil {
		fmt.Println("found 3")
	} else {
		fmt.Println("not found 3")
	}

	//同步到mysql
	mem.Sync()

}

//多主键
func multiKeyTest(rc *db.RedisClient) {
	if rc == nil {
		return
	}

	mem := db.NewMemMgr(rc, false, "test", "player", []string{"id", "name"})

	//查
	memMode := mem.QueryCheckDBByPKs(2, "ff")
	if memMode != nil {
		mem.AddData(memMode, false)
		id := memMode.Data["id"]
		name := memMode.Data["name"]
		info := memMode.Data["info"]
		test := memMode.Data["test"]
		fmt.Println("info query:", id, name, info, test)

		//改
		m := mem.UpdateByPks("info", "hihihi~", 2, "ff")
		if m != nil {
			id = m.Data["id"]
			name = m.Data["name"]
			info = m.Data["info"]
			test = m.Data["test"]
			fmt.Println("update query:", id, name, info, test)
			mem.SyncMemMode(m)
		}

		m = mem.UpdateByPks("test", "99.66", 2, "ff")
		if m != nil {
			id = m.Data["id"]
			name = m.Data["name"]
			info = m.Data["info"]
			test = m.Data["test"]
			fmt.Println("update query:", id, name, info, test)
			mem.SyncMemMode(m)
		}
	}

	//增
	memMode = db.NewMemMode()
	memMode.Data["id"] = 2
	memMode.Data["name"] = "kitty"
	memMode.Data["info"] = ",,,"
	memMode.Data["test"] = 23.4
	mem.AddData(memMode, false)
	mem.SyncMemMode(memMode)

	//查
	memMode = mem.QueryCheckDBByPK("2:kitty")
	if memMode != nil {
		id := memMode.Data["id"]
		name := memMode.Data["name"]
		info := memMode.Data["info"]
		test := memMode.Data["test"]
		fmt.Println("info query:", id, name, info, test)
	}

	//改
	m := mem.UpdateCheckDBByPks("info", "hello~", 2, "zz")
	if m != nil {
		id := m.Data["id"]
		name := m.Data["name"]
		info := m.Data["info"]
		test := m.Data["test"]
		fmt.Println("update query:", id, name, info, test)
		mem.SyncMemMode(m)
	}

	//删
	mem.DelDataCheckDBByPKs(2, "zz")
	memMode = mem.QueryCheckDBByPKs(2, "zz")
	if memMode != nil {
		fmt.Println("found 2")
	} else {
		fmt.Println("not found 2")
	}

}

func testMgr(dbMgr *db.MysqlMgr) {
	dbMgr.Query("user", func(rows *sql.Rows) {
		if rows == nil {
			return
		}

		columns, err := rows.Columns()
		if err != nil {
			fmt.Println(err)
			return
		}
		columnsLen := len(columns)

		scanArgs := make([]interface{}, columnsLen)
		values := make([]interface{}, columnsLen)
		for i := range values {
			scanArgs[i] = &values[i]
		}

		for rows.Next() {
			err := rows.Scan(scanArgs...)
			if err != nil {
				fmt.Println(err)
				break
			}

			for _, v := range values {
				d, ok := v.([]byte)
				if !ok {
					continue
				}
				fmt.Printf("%v ", string(d))
			}
			fmt.Println()
		}

	})
}

func redisTest(rc *db.RedisClient) {
	if rc == nil {
		return
	}
	rc.Set("a", 3, 5)
	//rc.Expire("a", 6)
	fmt.Println(rc.Get("a"))
}

func testORM(db *gorm.DB) {
	var users []*User
	err := db.Find(&users).Error
	if err == nil {
		for _, u := range users {
			fmt.Println(u)
		}
	}
	var user User
	err = db.First(&user).Error
	fmt.Println(user, err)

	var user1 User
	err = db.Last(&user1).Error
	fmt.Println(user1, err)

	var user2 User
	err = db.First(&user2, 2).Error
	fmt.Println(user2, err)

	var user3 User
	err = db.First(&user3, "name = ?", "Toney").Error
	fmt.Println(user3, err)

	var user4 User
	err = db.First(&user4, "id = ?", 2).Error
	fmt.Println(user4, err)

	var user5 User
	err = db.Where("id = ?", 2).First(&user5).Error
	fmt.Println(user5, err)

	err = db.Not("name", "Toney").Find(&users).Error
	for _, u := range users {
		fmt.Println(u)
	}

	var user6 User
	err = db.Not("name", "Jim").First(&user6).Error
	fmt.Println(user6, err)

	u1 := User{
		Name: "tom222",
		Remark: 1,
		Test: 222,
	}
	err = db.Create(&u1).Error
	fmt.Println(err)


	user5.Name = "jim777"
	err = db.Save(&user5).Error
	//err = db.Model(&user5).Update("name", "jim999").Error
	fmt.Println(err)


	//err = db.Delete(&user5).Error
	fmt.Println(err)

	err = db.Where("name like ?", "jim888").Delete(User{}).Error
	fmt.Println(err)

	fmt.Println("-----------------------------------------------------------")
}

func main() {
	dbMgr := db.NewMysqlMgr("test", "root", "123456", "localhost:3306")
	if dbMgr == nil {
		fmt.Println("create db mgr error")
		return
	}
	defer dbMgr.Close()

	testORM(dbMgr.DBConn())

	//testMgr(dbMgr)

	rc := db.NewRedisClient("localhost:6379", 3, 0)
	defer rc.Close()

	//
	//redisTest(rc)

	//
	//singleKeyTest(rc)

	//
	//multiKeyTest(rc)

}
