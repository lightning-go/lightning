/**
 * Created: 2020/4/2
 * @author: Jason
 */

package db

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strconv"
	"testing"
	"time"
)

type User struct {
	Id     int64   `json:"id"`
	Name   string  `json:"name"`
	Remark int     `json:"remark"`
	Test   float32 `json:"test"`
}

func (u *User) toString() string {
	return fmt.Sprintf("%d, %s, %d, %f", u.Id, u.Name, u.Remark, u.Test)
}


func newDBMgr() (*DBMgr, *RedisClient) {
	dbMgr := NewDBMgr(DB_type_mysql, "test", "root", "123456", "127.0.0.1:3306")
	rc := NewRedisClient("127.0.0.1:6379", 1, 0)
	return dbMgr, rc
}

func Test1(t *testing.T) {
	dbMgr, rc := newDBMgr()
	defer func (){
		dbMgr.Close()
		rc.Close()
	}()

	test(rc)
}

func test(rc *RedisClient) {
	memUser := NewMemMgr(rc, "test", "user", "id",
		func(obj interface{}) interface{} {
			u, ok := obj.(*User)
			if ok {
				return u.Id
			}
			return nil
		})

	ikName := "name"
	ikNameCallback := func(obj interface{})(string,interface{}){
		u, ok := obj.(*User)
		if ok {
			return ikName, u.Name
		}
		return "", nil
	}

	ikRemark := "name:remark"
	ikRemarkCallback := func(obj interface{})(string,interface{}){
		u, ok := obj.(*User)
		if ok {
			return ikRemark, fmt.Sprintf("%s:%d", u.Name, u.Remark)
		}
		return "", nil
	}

	memUser.SetIKCallback([]IkCallback {ikNameCallback, ikRemarkCallback})

	u := &User{}
	where := "name = 'jason1'"
	err := memUser.GetDataByIK(ikName, "jason1", u, where)
	if err == nil {
		fmt.Println("ikName get", u.toString())
	}

	u = &User{}
	err = memUser.GetData(1, u)
	if err == nil {
		fmt.Println("pk get", u.toString())
	}

	id := memUser.GetPKIncr()
	u = &User{}
	err = memUser.GetData(id, u)
	if err == gorm.ErrRecordNotFound {
		if id > 0 {
			remark, _ := strconv.Atoi(strconv.FormatInt(id, 10))
			u = &User{
				Id:     id,
				Name:   fmt.Sprintf("%s%d", "jason", id),
				Remark: remark,
				Test:   2323.2,
			}
			memUser.AddData(u)

			u3 := &User{}
			rn := fmt.Sprintf("%s:%d", u.Name, u.Remark)
			err = memUser.GetDataByIK(ikRemark, rn, u3)
			if err == nil {
				fmt.Println("ikRemark get", u3.toString())
			}

			u3 = &User{}
			err = memUser.GetDataByIK(ikName, u.Name, u3)
			if err == nil {
				fmt.Println("ikName get", u3.toString())
				memUser.CleanIkData(u3)
				u3.Name += fmt.Sprintf("update%d", u3.Id)
				memUser.UpdateData(u3)
				fmt.Println("ikName update", u3.toString())

				time.Sleep(time.Second * 3)
				memUser.DelData(u3)

				u4 := &User{}
				where = fmt.Sprintf("name = '%s'", u3.Name)
				err := memUser.GetDataByIK(ikName, u3.Name, u4, where)
				if err == nil {
					fmt.Println("ikName get", u4.toString())
				} else {
					fmt.Printf("ikName %s not found\n", u3.Name)
				}
			}
		}

	} else {
		fmt.Println("pk get", u.toString())
	}


}
