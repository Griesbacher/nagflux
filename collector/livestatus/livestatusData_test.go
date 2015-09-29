package livestatus

import (
	"fmt"
	"github.com/griesbacher/nagflux/helper"
	"reflect"
	"testing"
)

func TestLivestatusDataSanitizeValues(t *testing.T) {
	live1 := LivestatusData{"&", "host", "service", "comm\\ent a", "0", "author"}
	live2 := live1
	live2.sanitizeValues()
	if !reflect.DeepEqual(live1, live2) {
		t.Errorf("Sanitize should not change the comment. \n1:%s\n2:%s", live1, live2)
	}
}

func TestGetTablename(t *testing.T) {
	live := LivestatusData{"&", "host", "service", "comment", "0", "author"}
	tablename := fmt.Sprintf("%s%s%s%smessages", live.host_name, "&", live.service_display_name, "&")
	if live.getTablename() != tablename {
		t.Error("Tablename should match")
	}
	tablename2 := fmt.Sprintf("%s%s%s%smessages", live.host_name, "", live.service_display_name, "")
	if live.getTablename() == tablename2 {
		t.Error("Tablname should not match")
	}
}

func TestGenInfluxLineWithValue(t *testing.T) {
	live := LivestatusData{"&", "host", "service", "comment", "0", "author"}

	expected := fmt.Sprintf("%s%s value=\"%s\" %s", live.getTablename(), ",author="+live.author, "special text", helper.CastStringTimeFromSToMs(live.entry_time))
	result := live.genInfluxLineWithValue("", "special text")
	if expected != result {
		t.Errorf("Expected:%s\nResult:%s", expected, result)
	}
}

func TestGenInfluxLine(t *testing.T) {
	live := LivestatusData{"&", "host", "service", "comment", "0", "author"}
	expected := fmt.Sprintf("%s%s value=\"%s\" %s", live.getTablename(), ",a=1,b=2,author="+live.author, "comment", helper.CastStringTimeFromSToMs(live.entry_time))
	result := live.genInfluxLine(",a=1,b=2")
	if expected != result {
		t.Errorf("Expected:%s\nResult:%s", expected, result)
	}
}
