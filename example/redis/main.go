/**
 * Created: 2019/6/15
 * @author: Jason
 */

package main

import (
	"github.com/lightning-go/lightning/db"
	"fmt"
	"database/sql"
	"strconv"
)

type User struct {
	Id     int    `db:"id"`
	Name   string `db:"name"`
	Remark int    `db:"remark"`
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

func main() {
	dbMgr := db.NewMysqlMgr("test", "root", "123456", "localhost:3306")
	if dbMgr == nil {
		fmt.Println("create db mgr error")
		return
	}
	defer dbMgr.Close()

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
		}

	})

	pool := db.InitRedisPool("localhost:6379", 3, 0)
	if pool == nil {
		fmt.Println("init redis pool error")
		return
	}

	rc := db.NewRedisClient(pool)
	defer rc.Close()

	//
	singleKeyTest(rc)

	//
	multiKeyTest(rc)

}
