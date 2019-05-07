/**
 * Created: 2019/5/7 0007
 * @author: Jason
 */

package selector

import (
	"testing"
	"github.com/lightning-go/lightning/network"
	"fmt"
)

func TestSelector(t *testing.T) {
	s := genSessionData()
	if s == nil {
		fmt.Println("create failed")
		return
	}

	n := 0
	for i := 0; i < 20; i++ {
		session := s.SelectWeightLeast()
		fmt.Println(session.conn.GetId())

		if n == 5 {
			s.Update(session.conn.GetId(), 10)
			session := s.SelectWeightLeast()
			fmt.Println("update after:", session.conn.GetId())
		}
		n++
	}

}

func genSessionData() *WeightSelector {
	s := NewWeightSelector()

	for i := 1; i <= 10; i++ {
		conn := network.NewConnection(nil)
		if conn == nil {
			continue
		}

		sd := &SessionData{
			conn: conn,
			Host: "127.0.0.1:10000",
			Name: fmt.Sprintf("conn%v", i),
			Type: 1,
			Weight: 1,
		}

		s.Add(sd)
	}

	return s
}
