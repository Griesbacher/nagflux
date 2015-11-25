package livestatus

import (
	"testing"
)

var PrintCommentData = []struct {
	input  CommentData
	output string
}{
	{CommentData{Data: Data{fieldSeperator: "&", hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world"}, entryType: "1"},
		`host\ 1&service\ 1&messages,type=comment,author=philip value="hallo world" 000`},
	{CommentData{Data: Data{fieldSeperator: "&", hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world"}, entryType: "2"},
		`host\ 1&service\ 1&messages,type=downtime,author=philip value="hallo world" 000`},
	{CommentData{Data: Data{fieldSeperator: "&", hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world"}, entryType: "3"},
		`host\ 1&service\ 1&messages,type=flapping,author=philip value="hallo world" 000`},
	{CommentData{Data: Data{fieldSeperator: "&", hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world"}, entryType: "4"},
		`host\ 1&service\ 1&messages,type=acknowledgement,author=philip value="hallo world" 000`},
	{CommentData{Data: Data{fieldSeperator: "&", hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world"}, entryType: "5"},
		`host\ 1&service\ 1&messages,author=philip value="hallo world" 000`},
}

func TestSanitizeValuesComment(t *testing.T) {
	t.Parallel()
	comment := CommentData{Data: Data{fieldSeperator: "&", hostName: "host 1", serviceDisplayName: "service 1", author: "philip"}, entryType: "1"}
	comment.sanitizeValues()
	if comment.Data.hostName != `host\ 1` {
		t.Errorf("The notificationType should be escaped. Expected: %s Got: %s", `host\ 1`, comment.Data.hostName)
	}
}

func TestPrintComment(t *testing.T) {
	t.Parallel()
	comment := CommentData{Data: Data{fieldSeperator: "&", hostName: "host 1", serviceDisplayName: "service 1", author: "philip", comment: "hallo world"}, entryType: "1"}
	if !didThisPanic(comment.Print, 0.8) {
		t.Errorf("This should panic, due to unsuported influxdb version")
	}
	for _, data := range PrintCommentData {
		actual := data.input.Print(0.9)
		if actual != data.output {
			t.Errorf("Print(%s): expected: %s, actual: %s", data.input, data.output, actual)
		}
	}
}
