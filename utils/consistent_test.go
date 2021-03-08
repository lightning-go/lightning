
package utils

import (
	"fmt"
	"testing"
)

func calculate(c *ConsistentHash, users []string) {
	if c == nil || users == nil {
		return
	}
	fmt.Println("----------")

	d := make(map[string]int)
	for _, u := range users {
		server := c.GetNode(u)
		v, ok := d[server]
		if !ok {
			d[server] = 1
		} else {
			d[server] = v + 1
		}

	}
	for k, v := range d {
		fmt.Println("s:", k, ", num:", v)
	}
}


func TestExample2(t *testing.T) {
	users := make([]string, 0)
	for i := 0; i < 3000; i++ {
		v := i + 1
		users = append(users, fmt.Sprintf("%d", v))
	}

	c := NewConsistentHash()
	c.AddNodes([]string{
		"s1", "s2", "s3",
	})
	calculate(c, users)

	c.AddNodes([]string{
		"s4", "s5", "s6", "s7",
	})
	calculate(c, users)

	c.Remove("s5")
	calculate(c, users)

	c.RemoveNodes([]string{"s7", "s1", "s2"})
	calculate(c, users)

	c.RemoveNodes([]string{"s4", "s6"})
	calculate(c, users)

	c.AddNodes([]string{"s1", "s2"})
	calculate(c, users)

	c.Add("s4")
	calculate(c, users)

}
