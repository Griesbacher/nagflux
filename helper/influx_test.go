package helper

import "testing"

var SanitizeInfluxData = []struct {
	input  string
	output string
}{
	{"a\\a", "a\\\\a"},
	{"a a", "a\\ a"},
	{"a,a", "a\\,a"},
	{"\\, ", "\\\\\\,\\ "},
	{"aa", "aa"},
	{"", ""},
}

func TestSanitizeInfluxInput(t *testing.T) {
	for _, data := range SanitizeInfluxData {
		actual := SanitizeInfluxInput(data.input)
		if actual != data.output {
			t.Errorf("SanitizeInfluxData(%s): expected: %s, actual: %s", data.input, data.output, actual)
		}
	}
}
