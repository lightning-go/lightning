/**
 * Created: 2019/5/7 0007
 * @author: Jason
 */

package etcd

import (
	"fmt"
	"testing"
)

func TestEtcd(t *testing.T) {
	hosts := []string{"127.0.0.1:2379"}
	e := NewEtcd(hosts)
	if e == nil {
		fmt.Println("init etcd failed")
		return
	}

	e.Put("hi", "world")
	e.Get("hi", func(k, v []byte) {
		fmt.Println("get value:", string(k), string(v))
	})

	e.Put("/hello/1", "hi")
	e.Put("/hello/2", "hello")
	e.Get("/hello", func(k, v []byte) {
		fmt.Println("get value:", string(k), string(v))
	})

	//
	e.Put("/hi/1", "hi---")
	e.Put("/hi/2", "hi===")

	e.Watch("/hi", func(k, v []byte) {
		fmt.Println("watch put value:", string(k), string(v))
	}, func(k []byte) {
		fmt.Println("watch del value:", string(k))
	})

	e.Put("/hi/1", "hello---")
	e.Delete("/hi/2", func(k, v []byte) {
		fmt.Println("del value:", string(k), string(v))
	})

}
