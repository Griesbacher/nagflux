package helper

import (
	"errors"
	"net"
	"time"
)

func WaitForPort(typ, address string, timeout time.Duration) error {
	done := make(chan bool)
	timeOver := make(chan bool)
	go func() {
		for {
			select {
			case <-timeOver:
				return
			default:
				conn, err := net.Dial(typ, address)
				if err == nil {
					defer conn.Close()
					done <- true
					break
				}
			}
		}
	}()
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		timeOver <- false
		return errors.New("timeout")
	}
}
