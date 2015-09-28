package livestatus

import (
	"bufio"
	"log"
	"net"
	"testing"
	"github.com/griesbacher/nagflux/logging"
	"time"
	"reflect"
)

type MockLivestatus struct {
	LivestatusAddress string
	ConnectionType    string
	Queries           map[string]string
	isRunning         bool
}

func (mockLive *MockLivestatus)StartMockLivestatus() {
	var listener net.Listener
	var err error
	switch mockLive.ConnectionType {
	case "tcp":
		listener, err = net.Listen("tcp", mockLive.LivestatusAddress)
	case "file":
		listener, err = net.Listen("unix", mockLive.LivestatusAddress)
	default:
		log.Panic("ConnectionType undefined")
		return
	}

	if err != nil {
		log.Panic(err)
	}
	mockLive.isRunning = true
	for mockLive.isRunning {

		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		connReader := bufio.NewReader(conn)
		connWriter := bufio.NewWriter(conn)
		query := ""
		line, _ := connReader.ReadString('\n')
		for line != "\n" {
			query += line
			line, _ = connReader.ReadString('\n')
		}
		query += "\n"
		answer := mockLive.Queries[query]
		connWriter.WriteString(answer)
		connWriter.Flush()
		conn.Close()
	}
}

func (mockLive *MockLivestatus)StopMockLivestatus() {
	mockLive.isRunning = false
}

func TestConnectToLivestatus(t *testing.T) {
	//Create Livestatus mock
	address := "localhost:6557"
	typ := "tcp"
	livestatus := MockLivestatus{address, typ, map[string]string{"test\n\n":"foo;bar\n"}, true}

	go livestatus.StartMockLivestatus()
	connector := LivestatusConnector{logging.GetLogger(), livestatus.LivestatusAddress, livestatus.ConnectionType }

	csv := make(chan []string)
	finished := make(chan bool)
	go connector.connectToLivestatus("test\n\n", csv, finished)

	expected := []string{"foo", "bar"}

	waitingForTheEnd := true
	for waitingForTheEnd {
		select {
		case line := <-csv:
			if !reflect.DeepEqual(line, expected) {
				t.Errorf("Expected:%s result:%s", expected, line)
			}
		case result := <-finished:
			if !result {
				t.Error("Connector exited with error")
			}
			waitingForTheEnd = false
		case <-time.After(time.Duration(3)*time.Second):
			t.Error("Livestatus connection timed out")
		}
	}

}
