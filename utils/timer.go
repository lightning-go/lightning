/**
 * Created: 2019/4/24 0024
 * @author: Jason
 */

package utils

import "time"

type Timer struct {
	tick    time.Duration
	quit    chan struct{}
	f       func()
	running bool
}

func NewTimer(tick time.Duration, f func()) *Timer {
	return &Timer{
		tick: tick,
		f:    f,
		quit: make(chan struct{}),
	}
}

func (t *Timer) Start(once ...bool) {
	if t.f == nil {
		return
	}
	t.running = true

	go func() {
		tick := time.NewTicker(t.tick)
		defer tick.Stop()
	QUIT:
		for {
			select {
			case <-tick.C:
				t.f()
				if len(once) > 0 && once[0] {
					break QUIT
				}
			case <-t.quit:
				break QUIT
			}
		}
	}()
}


func (t *Timer) Stop() {
	t.running = false
	close(t.quit)
}

func (t *Timer) IsRunning() bool {
	return t.running
}



