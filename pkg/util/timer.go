package util

import (
	"os"
	"time"
)

func RepeatDo(step time.Duration, exit chan bool, signal chan os.Signal, fn func(at time.Time)) {
	ticker := time.NewTicker(step)
	go func() {
		for {
			select {
			case <-signal:
				return
			case <-exit:
				return
			case t := <-ticker.C:
				fn(t)
			}
		}
	}()
}
