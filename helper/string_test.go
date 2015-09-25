package helper

import (
	"reflect"
	"testing"
)

var StringToMapData = []struct {
	string           string
	entrySplitter    string
	keyValueSplitter string
	expected         map[string]string
}{
	{"k1=v1;k2=v2", ";", "=", map[string]string{"k1": "v1", "k2": "v2"}},
	{"k1=v1;k2=v2", ";", "", nil},
	{"k1=v1;k2=v2", "", "=", nil},
	{"", ";", "=", nil},
}

func TestStringToMap(t *testing.T) {
	for _, data := range StringToMapData {
		actual := StringToMap(data.string, data.entrySplitter, data.keyValueSplitter)
		if !reflect.DeepEqual(actual, data.expected) {
			t.Errorf("StringToMap(%s): expected:%s, actual:%s", data.string, data.expected, actual)
		}
	}
}

var StringIntToStringFloatData = []struct {
	input    string
	expected string
}{
	{"1.0", "1.0"},
	{"1", "1.0"},
	{"", ""},
}

func TestStringIntToStringFloat(t *testing.T) {
	for _, data := range StringIntToStringFloatData {
		actual := StringIntToStringFloat(data.input)
		if actual != data.expected {
			t.Errorf("StringIntToStringFloat(%s): expected:%s, actual:%s", data.input, data.expected, actual)
		}
	}
}

var CastStringTimeFromSToMsData = []struct {
	input    string
	expected string
}{
	{"1", "1000"},
	{"", "000"},
}

func TestCastStringTimeFromSToMs(t *testing.T) {
	for _, data := range CastStringTimeFromSToMsData {
		actual := CastStringTimeFromSToMs(data.input)
		if actual != data.expected {
			t.Errorf("CastStringTimeFromSToMs(%s): expected:%s, actual:%s", data.input, data.expected, actual)
		}
	}
}
