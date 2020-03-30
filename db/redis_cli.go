/**
 * @author: Jason
 * Created: 19-5-3
 */

package db

import (
	"github.com/gomodule/redigo/redis"
	"github.com/lightning-go/lightning/conf"
	"time"
)

func InitRedisPool(addr string, maxIdle, maxActive int) *redis.Pool {
	return &redis.Pool{
		MaxIdle: maxIdle,
		MaxActive: maxActive,
		IdleTimeout: conf.GetGlobalVal().RedisIdleTimeout,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", addr)
			if err != nil {
				panic(err)
			}
			return conn, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

type RedisClient struct {
	pool *redis.Pool
}

func NewRedisClient(pool *redis.Pool) *RedisClient {
	if pool == nil {
		return nil
	}
	rc := &RedisClient{
		pool: pool,
	}
	return rc
}

func (rc *RedisClient) Close() error {
	return rc.pool.Close()
}

func (rc *RedisClient) GetConn() redis.Conn {
	return rc.pool.Get()
}

func (rc *RedisClient) CloseConn(conn redis.Conn) {
	if conn == nil {
		return
	}
	conn.Close()
}

func (rc *RedisClient) Expire(key string, second int64) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	d, err := conn.Do("EXPIRE", key, second)
	return d, err
}

func (rc *RedisClient) Get(key string) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	d, err := conn.Do("GET", key)
	return d, err
}

func (rc *RedisClient) MGet(keys ...interface{}) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	d, err := redis.Strings(conn.Do("MGET", keys...))
	return d, err
}

func (rc *RedisClient) HGet(keys ...interface{}) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	d, err := conn.Do("HGET", keys...)
	return d, err
}

func (rc *RedisClient) HMGet(keys ...interface{}) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	d, err := redis.Strings(conn.Do("HMGET", keys...))
	return d, err
}

func (rc *RedisClient) HGetAll(key string) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	d, err := conn.Do("HGETALL", key)
	return d, err
}

func (rc *RedisClient) Set(key string, v interface{}, expire ...int64) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	var d interface{}
	var err error
	if len(expire) > 0 {
		d, err = conn.Do("SET", key, v, "EX", expire[0])
	} else {
		d, err = conn.Do("SET", key, v)
	}
	return d, err
}

func (rc *RedisClient) MSet(kvs ...interface{}) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	d, err := conn.Do("MSET", kvs...)
	return d, err
}

func (rc *RedisClient) HSet(v ...interface{}) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	d, err := conn.Do("HSET", v...)
	return d, err
}

func (rc *RedisClient) Incr(key string) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	d, err := conn.Do("INCR", key)
	return d, err
}

func (rc *RedisClient) Del(key string) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	d, err := conn.Do("DEL", key)
	return d, err
}

func (rc *RedisClient) HDel(v ...interface{}) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	d, err := conn.Do("HDEL", v...)
	return d, err
}

func (rc *RedisClient) PipeSet(conn redis.Conn, key, value interface{}) {
	conn.Send("SET", key, value)
}

func (rc *RedisClient) PipeHSet(conn redis.Conn, key, field, value interface{}) {
	conn.Send("HSET", key, field, value)
}

func (rc *RedisClient) PipeZAdd(conn redis.Conn, v ...interface{}) {
	conn.Send("ZADD", v...)
}

func (rc *RedisClient) PipeHGet(conn redis.Conn, key, field interface{}) {
	conn.Send("HGET", key, field)
}

func (rc *RedisClient) PipeHDel(conn redis.Conn, v ...interface{}) {
	conn.Send("HDEL", v...)
}

func (rc *RedisClient) PipeEnd(conn redis.Conn) {
	conn.Flush()
}

func (rc *RedisClient) PipeRecv(conn redis.Conn) (interface{}, error) {
	return conn.Receive()
}
