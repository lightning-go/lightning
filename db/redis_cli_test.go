package db

import (
	"fmt"
	"testing"
	"time"
)


var (
	redisAddr = "127.0.0.1:6379"
	redisMaxIdle = 1
	redisMaxActive = 0
)


func TestSet(t *testing.T) {
	rc := NewRedisClient(redisAddr, redisMaxIdle, redisMaxActive)
	defer rc.Close()

	key := "test1"
	rc.Set(key, 333333333)

	v, _ := rc.Get(key)
	fmt.Println(key, string(v.([]uint8)))


	rc.Set(key, 5555, 10)
	v, _ = rc.Get(key)
	fmt.Println(key, string(v.([]uint8)))

	time.Sleep(10 * time.Second)
	v, _ = rc.Get(key)
	fmt.Println(key, v)

}

func TestHSet(t *testing.T) {
	rc := NewRedisClient(redisAddr, redisMaxIdle, redisMaxActive)
	defer rc.Close()

	key := "test3"
	field := "testField"
	for i := 1; i < 10; i++ {
		rc.HSet(key, fmt.Sprintf("%v%v", field, i), i)
	}

	for i := 1; i < 10; i++ {
		d, err := rc.HGet(key, fmt.Sprintf("%v%v", field, i))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(d.([]uint8)))
	}

}

func TestZAdd(t *testing.T) {
	rc := NewRedisClient(redisAddr, redisMaxIdle, redisMaxActive)
	defer rc.Close()

	key := "test2"
	v, err := rc.ZAdd(key, map[string]string{
		"hi": "3",
		"hello": "8",
		"jjjj": "9",
		"eeee": "3",
		"pppp": "5",
		"nnnn": "2",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(key, v)

	d, err := rc.ZRange(key, 0, -1)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(key, d)

}

func TestPip(t *testing.T) {
	rc := NewRedisClient(redisAddr, redisMaxIdle, redisMaxActive)
	defer rc.Close()

	test4 := "test4"
	test5 := "test5"
	test6 := "test6"
	test7 := "test7"
	conn := rc.GetConn()
	rc.PipeSet(conn, test4, test4)
	rc.PipeSet(conn, test5, test5)
	rc.PipeSet(conn, test6, test6)
	rc.PipeSet(conn, test7, test7)
	rc.PipeEnd(conn)
	d, err := rc.PipeRecv(conn)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("pip set", d)
	rc.CloseConn(conn)

	d, err = rc.MGet(test4, test5, test6, test7)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("pip", d)

	conn = rc.GetConn()
	rc.PipeDel(conn, test4)
	rc.PipeDel(conn, test5)
	rc.PipeDel(conn, test7)
	rc.PipeEnd(conn)
	d, err = rc.PipeRecv(conn)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("pip del", d)
	rc.CloseConn(conn)
}
