package helper

import (
	"reflect"
	"testing"
)

func TestCopyMap(t *testing.T) {
	map1 := map[string]string{"k1": "v1", "k2": "v2"}
	map2 := CopyMap(map1)

	if !reflect.DeepEqual(map1, map2) {
		t.Error("Maps are not equal after copy.")
	}

	map2["k1"] = "foo"

	if reflect.DeepEqual(map1, map2) {
		t.Error("Maps are equal after change.")
	}
}

func TestPrintMapAsString(t *testing.T) {
	t.Parallel()
	map1 := map[string]string{"k1": "v1", "k2": "v2"}
	result := PrintMapAsString(map1, ";", "=")
	expected1 := "k1=v1;k2=v2"
	expected2 := "k2=v2;k1=v1"
	if result != expected1 && result != expected2 {
		t.Errorf("Failed: PrintMapAsString() expected:%s/%s result:%s", expected1, expected2, result)
	}

	map1 = map[string]string{"k1": "v1", "k2": "v2"}
	result = PrintMapAsString(map1, "", "")
	expected1 = "k1v1k2v2"
	expected2 = "k2v2k1v1"
	if result != expected1 && result != expected2 {
		t.Errorf("Failed: PrintMapAsString() expected:%s/%s result:%s", expected1, expected2, result)
	}

}
