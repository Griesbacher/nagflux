package helper

import "testing"

var CreateJSONFromStringMapData = []struct {
	input    map[string]string
	expected string
}{
	{map[string]string{"a": "1"}, `,"a":1`},
	{map[string]string{"a": "b"}, `,"a":"b"`},
	{map[string]string{"a": "1", "2": "b"}, `,"a":1,2:"b"`},
}

func TestCreateJSONFromStringMap(t *testing.T) {
	t.Parallel()
	for _, data := range CreateJSONFromStringMapData {
		actual := CreateJSONFromStringMap(data.input)
		if actual != data.expected {
			t.Errorf("CreateJSONFromStringMap(%s): expected:%s, actual:%s", data.input, data.expected, actual)
		}
	}
}

var SanitizeElasicInputData = []struct {
	input    string
	expected string
}{
	{"asdf", "asdf"},
	{"'asdf'", "asdf"},
	{"'as df'", "as df"},
	{`'as\ df'`, `as\\ df`},
	{`'as\" df'`, `as\\\" df`},
}

func TestSanitizeElasicInput(t *testing.T) {
	t.Parallel()
	for _, data := range SanitizeElasicInputData {
		actual := SanitizeElasicInput(data.input)
		if actual != data.expected {
			t.Errorf("SanitizeElasicInputData(%s): expected:%s, actual:%s", data.input, data.expected, actual)
		}
	}
}
