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
	{"k1=v1;k2=", ";", "=", map[string]string{"k1": "v1"}},
	{"k1=v1;k2=v2", ";", "", map[string]string{}},
	{"k1=v1;k2=v2", "", "=", map[string]string{}},
	{"", ";", "=", map[string]string{}},
}

func TestStringToMap(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	for _, data := range CastStringTimeFromSToMsData {
		actual := CastStringTimeFromSToMs(data.input)
		if actual != data.expected {
			t.Errorf("CastStringTimeFromSToMs(%s): expected:%s, actual:%s", data.input, data.expected, actual)
		}
	}
}

var IsStringANumberData = []struct {
	input    string
	expected bool
}{
	{"1", true},
	{"1.0", true},
	{"1,0", false},
	{"a", false},
}

func TestIsStringANumber(t *testing.T) {
	t.Parallel()
	for _, data := range IsStringANumberData {
		actual := IsStringANumber(data.input)
		if actual != data.expected {
			t.Errorf("IsStringANumber(%s): expected:%t, actual:%t", data.input, data.expected, actual)
		}
	}
}

var VersionOrdinalData = []struct {
	input    string
	input2   string
	expected bool
}{
	{"1", "2", true},
	{"2", "1", false},
	{"1", "1", false},
	{"a", "b", true},
	{"b", "a", false},
	{"0", "0", false},
	{".", ",", false},
	{"1.10", "1.09", false},
	{"1.1.10", "1.1.09", false},
	{"1.09", "1.10", true},
	{"1.1.09", "1.1.10", true},
}

func TestVersionOrdinal(t *testing.T) {
	t.Parallel()
	for _, data := range VersionOrdinalData {
		actual := VersionOrdinal(data.input) < VersionOrdinal(data.input2)
		if actual != data.expected {
			t.Errorf("VersionOrdinal(%s) < VersionOrdinal(%s): expected:%t, actual:%t", data.input, data.input2, data.expected, actual)
		}
	}
}
