package utils

import (
	"github.com/lightning-go/lightning/logger"
	"testing"
	"time"
)



func TestTimer(t *testing.T) {
	timer := NewHeapTimer(100)

	timer.AddTimer(func() {
		logger.Trace("timer1...5")
	}, 5000, true)

	timer.AddTimer(func() {
		logger.Trace("timer2...2.1")
	}, 2000, false)

	timer.Run()

	timer.AddTimer(func() {
		logger.Trace("timer3...2")
	}, 2000, true)

	timer.AddTimer(func() {
		logger.Trace("timer4...3")
	}, 3000, true)

	timer.AddTimer(func() {
		logger.Trace("timer5...3.1")
	}, 3000, false)

	time.Sleep(time.Second * 20)
	timer.Stop()

	time.Sleep(time.Second * 3)
}
