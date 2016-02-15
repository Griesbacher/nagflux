package livestatus

import (
	"testing"
)

var PrintCommentData = []struct {
	input  CommentData
	output string
}{
	{CommentData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world"}, entryType: "1"},
		`messages,host=host\ 1,service=service\ 1,type=comment,author=philip message="hallo world" 000`},
	{CommentData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world"}, entryType: "2"},
		`messages,host=host\ 1,service=service\ 1,type=downtime,author=philip message="hallo world" 000`},
	{CommentData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world"}, entryType: "3"},
		`messages,host=host\ 1,service=service\ 1,type=flapping,author=philip message="hallo world" 000`},
	{CommentData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world"}, entryType: "4"},
		`messages,host=host\ 1,service=service\ 1,type=acknowledgement,author=philip message="hallo world" 000`},
	{CommentData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world"}, entryType: "5"},
		`messages,host=host\ 1,service=service\ 1,author=philip message="hallo world" 000`},
}

func TestSanitizeValuesComment(t *testing.T) {
	t.Parallel()
	comment := CommentData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip"}, entryType: "1"}
	comment.sanitizeValues()
	if comment.Data.hostName != `host\ 1` {
		t.Errorf("The notificationType should be escaped. Expected: %s Got: %s", `host\ 1`, comment.Data.hostName)
	}
}

func TestPrintComment(t *testing.T) {
	t.Parallel()
	comment := CommentData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world"}, entryType: "1"}
	if !didThisPanic(comment.PrintForInfluxDB, 0.8) {
		t.Errorf("This should panic, due to unsuported influxdb version")
	}
	for _, data := range PrintCommentData {
		actual := data.input.PrintForInfluxDB(0.9)
		if actual != data.output {
			t.Errorf("Print(%s): expected: %s, actual: %s", data.input, data.output, actual)
		}
	}
}
