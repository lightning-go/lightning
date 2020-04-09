package db

import (
	"fmt"
	"strconv"
	"testing"
	"database/sql"
)

func singleKeyTest(rc *RedisClient) {
	if rc == nil {
		return
	}

	mem := NewMemMgr(rc, false, "test", "user", []string{"id"}, "name")
	mem.LoadDB()

	memMode := mem.QueryCheckDBByPK(strconv.Itoa(2))
	if memMode != nil {
		id := memMode.Data["id"]
		name := memMode.Data["name"]
		remark := memMode.Data["remark"]
		test := memMode.Data["test"]
		fmt.Println("query:", id, name, remark, test)

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

	memMode = NewMemMode()
	memMode.Data["id"] = 3
	memMode.Data["name"] = "Jason"
	memMode.Data["remark"] = 1
	memMode.Data["test"] = 5.01
	mem.AddData(memMode, false)
	mem.SyncMemMode(memMode)

	memMode = NewMemMode()
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

	mem.DelDataCheckDBByPK("3")
	memMode = mem.QueryCheckDBByPK(strconv.Itoa(3))
	if memMode != nil {
		fmt.Println("found 3")
	} else {
		fmt.Println("not found 3")
	}

	mem.Sync()
}


func multiKeyTest(rc *RedisClient) {
	if rc == nil {
		return
	}

	mem := NewMemMgr(rc, false, "test", "player", []string{"id", "name"})

	memMode := mem.QueryCheckDBByPKs(2, "ff")
	if memMode != nil {
		mem.AddData(memMode, false)
		id := memMode.Data["id"]
		name := memMode.Data["name"]
		info := memMode.Data["info"]
		test := memMode.Data["test"]
		fmt.Println("info query:", id, name, info, test)

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

	memMode = NewMemMode()
	memMode.Data["id"] = 2
	memMode.Data["name"] = "kitty"
	memMode.Data["info"] = ",,,"
	memMode.Data["test"] = 23.4
	mem.AddData(memMode, false)
	mem.SyncMemMode(memMode)

	memMode = mem.QueryCheckDBByPK("2:kitty")
	if memMode != nil {
		id := memMode.Data["id"]
		name := memMode.Data["name"]
		info := memMode.Data["info"]
		test := memMode.Data["test"]
		fmt.Println("info query:", id, name, info, test)
	}

	m := mem.UpdateCheckDBByPks("info", "hello~", 2, "zz")
	if m != nil {
		id := m.Data["id"]
		name := m.Data["name"]
		info := m.Data["info"]
		test := m.Data["test"]
		fmt.Println("update query:", id, name, info, test)
		mem.SyncMemMode(m)
	}

	mem.DelDataCheckDBByPKs(2, "zz")
	memMode = mem.QueryCheckDBByPKs(2, "zz")
	if memMode != nil {
		fmt.Println("found 2")
	} else {
		fmt.Println("not found 2")
	}

}

func testMgr(dbMgr *DBMgr) {
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

func TestSingle(t *testing.T) {
	dbMgr := NewDBMgr(DB_type_mysql, "test", "root", "123456", "localhost:3306")
	if dbMgr == nil {
		fmt.Println("create db mgr error")
		return
	}
	defer dbMgr.Close()

	testMgr(dbMgr)

	rc := NewRedisClient("localhost:6379", 3, 0)
	defer rc.Close()

	singleKeyTest(rc)
}

func TestMulti(t *testing.T) {
	dbMgr := NewDBMgr(DB_type_mysql, "test", "root", "123456", "localhost:3306")
	if dbMgr == nil {
		fmt.Println("create db mgr error")
		return
	}
	defer dbMgr.Close()

	rc := NewRedisClient("localhost:6379", 3, 0)
	defer rc.Close()

	multiKeyTest(rc)
}

func TestR(t *testing.T) {
	rc := NewRedisClient("localhost:6379", 3, 0)
	defer rc.Close()

	rc.Set("a", 3, 5)
	//rc.Expire("a", 6)
	fmt.Println(rc.Get("a"))
}
