package helper

import (
	"github.com/griesbacher/nagflux/config"
	"reflect"
	"testing"
)

var SanitizeInfluxData = []struct {
	input  string
	output string
}{
	{"a a", `a\ a`},
	{"a,a", `a\,a`},
	{", ", `\,\ `},
	{"aa", "aa"},
	{`c:\ `, `c:\\ `},
	{"", ""},
	{`"a a"`, `"a a"`},
	{`§`, `SS`},
}

var SanitizeInfluxDataMap = []struct {
	input  map[string]string
	output map[string]string
}{
	{map[string]string{"a a": "'asdf'"}, map[string]string{`a\ a`: "asdf"}},
	{map[string]string{"": "a,a"}, map[string]string{"": `a\,a`}},
	{map[string]string{", ": "aa"}, map[string]string{`\,\ `: "aa"}},
	{map[string]string{`c:\ `: ""}, map[string]string{`c:\\ `: ""}},
	{map[string]string{"": ""}, map[string]string{"": ""}},
}

func TestSanitizeInfluxInput(t *testing.T) {
	//t.Parallel()
	config.InitConfigFromString(`[Influx]
    Enabled = true
    Version = 0.9
    Address = "http://127.0.0.1:8086"
    Arguments = "precision=ms&u=root&p=root&db=nagflux"
    CreateDatabaseIfNotExists = true
    # leave empty to disable
    NastyString = "§"
    NastyStringToReplace = "SS"
    HostcheckAlias = "hostcheck"
`)
	for _, data := range SanitizeInfluxData {
		actual := SanitizeInfluxInput(data.input)
		if actual != data.output {
			t.Errorf("SanitizeInfluxData(%s): expected: %s, actual: %s", data.input, data.output, actual)
		}
	}
}

func TestSanitizeMap(t *testing.T) {
	t.Parallel()
	for _, data := range SanitizeInfluxDataMap {
		actual := SanitizeMap(data.input)
		if !reflect.DeepEqual(actual, data.output) {
			t.Errorf("SanitizeInfluxData(%s): expected: %s, actual: %s", data.input, data.output, actual)
		}
	}
}
