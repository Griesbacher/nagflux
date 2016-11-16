package livestatus

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/kdar/factorlog"
	"io"
	"net"
	"strings"
)

//Connector fetches data from livestatus.
type Connector struct {
	Log               *factorlog.FactorLog
	LivestatusAddress string
	ConnectionType    string
}

//Queries livestatus and returns an list of list outer list are lines inner elements within the line.
func (connector Connector) connectToLivestatus(query string, result chan []string, outerFinish chan bool) {
	var conn net.Conn
	switch connector.ConnectionType {
	case "tcp":
		conn, _ = net.Dial("tcp", connector.LivestatusAddress)
	case "file":
		conn, _ = net.Dial("unix", connector.LivestatusAddress)
	default:
		connector.Log.Critical("Connection type is unknown, options are: tcp, file. Input:" + connector.ConnectionType)
		outerFinish <- false
		return
	}
	if conn == nil {
		outerFinish <- false
		return
	}

	defer conn.Close()
	fmt.Fprintf(conn, query)
	reader := bufio.NewReader(conn)

	length := 1
	for length > 0 {
		message, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				connector.Log.Warn(err)
			}
		}
		length = len(message)
		if length > 0 {
			csvReader := csv.NewReader(strings.NewReader(string(message)))
			csvReader.Comma = ';'
			csvReader.LazyQuotes = true
			records, err := csvReader.Read()
			if err != nil {
				connector.Log.Warn("Query failed while csv parsing:" + query)
				connector.Log.Warn(string(message))
				connector.Log.Warn(err)
			}
			result <- records
		}
	}
	outerFinish <- true
}
