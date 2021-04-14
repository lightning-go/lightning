/**
 * @author: Jason
 * Created: 19-5-3
 */

package db

import (
	"github.com/gomodule/redigo/redis"
	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/logger"
	"github.com/mna/redisc"
	"time"
)

func createRedisPool(addr string, maxIdle, maxActive int, opts ...redis.DialOption) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: conf.GetGlobalVal().RedisIdleTimeout,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr, opts...)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

type IRedisClient interface {
	Get() redis.Conn
	Close() error
}

type RedisClient struct {
	pool IRedisClient
}

func NewRedisClusterClient(hostList []string, maxIdle, maxActive int) *RedisClient {
	cluster := &redisc.Cluster{
		StartupNodes: hostList,
		DialOptions:  []redis.DialOption{redis.DialConnectTimeout(5 * time.Second)},
		CreatePool: func(addr string, opts ...redis.DialOption) (*redis.Pool, error) {
			return createRedisPool(addr, maxIdle, maxActive, opts...), nil
		},
	}

	err := cluster.Refresh()
	if err != nil {
		logger.Errorf("Refresh failed: %v", err)
	}

	rc := &RedisClient{
		pool: cluster,
	}
	return rc
}

func NewRedisClient(host string, maxIdle, maxActive int) *RedisClient {
	pool := createRedisPool(host, maxIdle, maxActive)
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

func (rc *RedisClient) Expire(key string, second int64) (d interface{}, err error) {
	conn := rc.pool.Get()
	defer conn.Close()
	return conn.Do("EXPIRE", key, second)
}

func (rc *RedisClient) Get(key string) (d interface{}, err error) {
	conn := rc.pool.Get()
	defer conn.Close()
	return conn.Do("GET", key)
}

func (rc *RedisClient) MGet(keys ...interface{}) (d []string, err error) {
	conn := rc.pool.Get()
	defer conn.Close()
	return redis.Strings(conn.Do("MGET", keys...))
}

func (rc *RedisClient) HGet(key, field interface{}) (d interface{}, err error) {
	conn := rc.pool.Get()
	defer conn.Close()
	return conn.Do("HGET", key, field)
}

func (rc *RedisClient) HMGet(keys ...interface{}) (d []string, err error) {
	conn := rc.pool.Get()
	defer conn.Close()
	return redis.Strings(conn.Do("HMGET", keys...))
}

func (rc *RedisClient) HGetAll(key string) (d interface{}, err error) {
	conn := rc.pool.Get()
	defer conn.Close()
	return conn.Do("HGETALL", key)
}

func (rc *RedisClient) Set(key string, v interface{}, expire ...int64) (d interface{}, err error) {
	conn := rc.pool.Get()
	defer conn.Close()
	if len(expire) > 0 {
		d, err = conn.Do("SET", key, v, "EX", expire[0])
	} else {
		d, err = conn.Do("SET", key, v)
	}
	return d, err
}

func (rc *RedisClient) MSet(kvs ...interface{}) (d interface{}, err error) {
	conn := rc.pool.Get()
	defer conn.Close()
	return conn.Do("MSET", kvs...)
}

func (rc *RedisClient) HSet(key, field, v interface{}) (d interface{}, err error) {
	conn := rc.pool.Get()
	defer conn.Close()
	return conn.Do("HSET", key, field, v)
}

func (rc *RedisClient) Incr(key string) (d interface{}, err error) {
	conn := rc.pool.Get()
	defer conn.Close()
	return conn.Do("INCR", key)
}

func (rc *RedisClient) Del(key string) (d interface{}, err error) {
	conn := rc.pool.Get()
	defer conn.Close()
	return conn.Do("DEL", key)
}

func (rc *RedisClient) HDel(key, field interface{}) (d interface{}, err error) {
	conn := rc.pool.Get()
	defer conn.Close()
	return conn.Do("HDEL", key, field)
}

func (rc *RedisClient) ZAdd(key string, kvs map[string]string) (d interface{}, err error) {
	if kvs == nil || len(kvs) == 0 {
		return
	}
	conn := rc.pool.Get()
	defer conn.Close()

	str := []interface{}{key}
	for k, v := range kvs {
		str = append(str, v, k)
	}
	return conn.Do("ZADD", str...)
}

func (rc *RedisClient) ZRange(key string, start, end int) (d [][]string, err error) {
	conn := rc.pool.Get()
	defer conn.Close()
	var v interface{}
	v, err = conn.Do("ZRANGE", key, start, end, "WITHSCORES")
	if err != nil {
		return nil, err
	}
	return rc.parseZRangeWithScores(v), nil
}

func (rc *RedisClient) parseZRangeWithScores(v interface{}) [][]string {
	value, ok := v.([]interface{})
	if !ok {
		return nil
	}
	valueLen := len(value)
	d := make([][]string, 0)
	for i := 0; i < valueLen; i += 2 {
		v1, err := redis.String(value[i], nil)
		if err != nil {
			continue
		}
		if i + 1 < valueLen {
			v2, err := redis.String(value[i + 1], nil)
			if err != nil {
				continue
			}
			d = append(d, []string{string(v2), string(v1)})
		}
	}
	return d
}

func (rc *RedisClient) PipeSet(conn redis.Conn, key, value interface{}) {
	conn.Send("SET", key, value)
}

func (rc *RedisClient) PipeHSet(conn redis.Conn, key, field, value interface{}) {
	conn.Send("HSET", key, field, value)
}

func (rc *RedisClient) PipeZAdd(conn redis.Conn, key string, kvs map[string]string) {
	if kvs == nil || len(kvs) == 0 {
		return
	}
	str := []interface{}{key}
	for k, v := range kvs {
		str = append(str, v, k)
	}
	conn.Send("ZADD", str...)
}

func (rc *RedisClient) PipeHGet(conn redis.Conn, key, field interface{}) {
	conn.Send("HGET", key, field)
}

func (rc *RedisClient) PipeDel(conn redis.Conn, keys ...interface{}) {
	conn.Send("DEL", keys...)
}

func (rc *RedisClient) PipeHDel(conn redis.Conn, key, field interface{}) {
	conn.Send("HDEL", key, field)
}

func (rc *RedisClient) PipeEnd(conn redis.Conn) {
	conn.Flush()
}

func (rc *RedisClient) PipeRecv(conn redis.Conn) (interface{}, error) {
	return conn.Receive()
}
