/**
 * Created: 2019/5/7 0007
 * @author: Jason
 */

package selector

import (
	"testing"
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
		fmt.Println(session.Id)

		if n == 5 {
			s.Update(session.Id, 10)
			session := s.SelectWeightLeast()
			fmt.Println("update after:", session.Id)
		}
		n++
	}

}

func genSessionData() *WeightSelector {
	s := NewWeightSelector()

	for i := 1; i <= 10; i++ {
		sd := &SessionData{
			Id:     i,
			Host:   "127.0.0.1:10000",
			Name:   fmt.Sprintf("conn%v", i),
			Type:   1,
			Weight: 1,
		}

		s.Add(sd)
	}

	return s
}
