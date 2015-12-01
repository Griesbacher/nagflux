package helper

import (
	"net"
	"testing"
	"time"
)

const (
	address = "localhost:9999"
	typ     = "tcp"
)

func dummyServer() {
	l, err := net.Listen(typ, address)
	if err != nil {
		panic(err)
	}
	conn, err := l.Accept()
	if err != nil {
		panic(err)
	}
	conn.Close()
	l.Close()
}

func TestWaitForPort(t *testing.T) {
	t.Parallel()
	//timeout
	if err := WaitForPort(typ, address, time.Duration(100)*time.Millisecond); err == nil {
		t.Errorf("on %s %s should no service listen", typ, address)
	}
	go dummyServer()
	if err := WaitForPort(typ, address, time.Duration(2000)*time.Millisecond); err != nil {
		t.Errorf("on %s %s a service is listening", typ, address)
	}
}
