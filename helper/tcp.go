package helper

import (
	"errors"
	"net"
	"time"
)

//WaitForPort tries to connect to a server of returns with an error if no connection was made within the time
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
