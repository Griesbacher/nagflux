package livestatus

import (
	"testing"
	"github.com/griesbacher/nagflux/config"
	"fmt"
	"github.com/griesbacher/nagflux/logging"
)

var PrintCommentData = []struct {
	input        CommentData
	outputInflux string
	outputElastic string
}{
	{CommentData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world", entryTime:"1458988932000"}, entryType: "1"},
		`messages,host=host\ 1,service=service\ 1,type=comment,author=philip message="hallo world" 1458988932000000`,
		`{"index":{"_index":"index-2016.03","_type":"messages"}}
{"timestamp":1458988932000000,"message":"hallo world","author":"philip","host":"host 1","service":"service 1","type":"comment"}
`},
	{CommentData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world", entryTime:"1458988932000"}, entryType: "2"},
		`messages,host=host\ 1,service=service\ 1,type=downtime,author=philip message="hallo world" 1458988932000000`,
		`{"index":{"_index":"index-2016.03","_type":"messages"}}
{"timestamp":1458988932000000,"message":"hallo world","author":"philip","host":"host 1","service":"service 1","type":"downtime"}
`},
	{CommentData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world", entryTime:"1458988932000"}, entryType: "3"},
		`messages,host=host\ 1,service=service\ 1,type=flapping,author=philip message="hallo world" 1458988932000000`,
		`{"index":{"_index":"index-2016.03","_type":"messages"}}
{"timestamp":1458988932000000,"message":"hallo world","author":"philip","host":"host 1","service":"service 1","type":"flapping"}
`},
	{CommentData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world", entryTime:"1458988932000"}, entryType: "4"},
		`messages,host=host\ 1,service=service\ 1,type=acknowledgement,author=philip message="hallo world" 1458988932000000`,
		`{"index":{"_index":"index-2016.03","_type":"messages"}}
{"timestamp":1458988932000000,"message":"hallo world","author":"philip","host":"host 1","service":"service 1","type":"acknowledgement"}
`},
	{CommentData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world", entryTime:"1458988932000"}, entryType: "5"},
		`messages,host=host\ 1,service=service\ 1,author=philip message="hallo world" 1458988932000000`,
		`{"index":{"_index":"index-2016.03","_type":"messages"}}
{"timestamp":1458988932000000,"message":"hallo world","author":"philip","host":"host 1","service":"service 1","type":""}
`},
}

func TestSanitizeValuesComment(t *testing.T) {
	t.Parallel()
	comment := CommentData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip"}, entryType: "1"}
	comment.sanitizeValues()
	if comment.Data.hostName != `host\ 1` {
		t.Errorf("The notificationType should be escaped. Expected: %s Got: %s", `host\ 1`, comment.Data.hostName)
	}
}

func TestPrintInfluxdbComment(t *testing.T) {
	t.Parallel()
	logging.InitTestLogger()
	comment := CommentData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world"}, entryType: "1"}
	if !didThisPanic(comment.PrintForInfluxDB, "0.8") {
		t.Errorf("This should panic, due to unsuported influxdb version")
	}
	for _, data := range PrintCommentData {
		actual := data.input.PrintForInfluxDB("0.9")
		if actual != data.outputInflux {
			t.Errorf("Print(%s): expected: %s, actual: %s", data.input, data.outputInflux, actual)
		}
	}
}

func TestPrintElasticsearchComment(t *testing.T) {
	logging.InitTestLogger()
	config.InitConfigFromString(fmt.Sprintf(Config, "monthly"))
	comment := CommentData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world", entryTime:"1458988932000"}, entryType: "1"}
	if !didThatPanic(comment.PrintForElasticsearch, "1.0", "index") {
		t.Errorf("This should panic, due to unsuported elasticsearch version")
	}
	for _, data := range PrintCommentData {
		actual := data.input.PrintForElasticsearch("2.0", "index")
		if actual != data.outputElastic {
			t.Errorf("Print(%s): expected: %s, actual: %s", data.input, data.outputElastic, actual)
		}
	}
}