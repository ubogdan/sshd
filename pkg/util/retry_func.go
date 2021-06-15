package util

import (
	"errors"
	"fmt"
)

var ErrInterrupt = errors.New("用户退出")

func RetryFunc(times int, fn func() error) (err error) {
	for i := 0; i < times; i++ {
		if err = fn(); err == nil {
			return nil
		} else if errors.Is(err, ErrInterrupt) {
			return err
		}
	}
	return fmt.Errorf("重试%d次失败:%v", times, err)
}
