package helper

import (
	"reflect"
	"testing"
)

func TestCopyMap(t *testing.T) {
	map1 := map[string]string{"k1": "v1", "k2": "v2"}
	map2 := CopyMap(map1)

	eq := reflect.DeepEqual(map1, map2)
	if !eq {
		t.Error("Maps are not equal after copy.")
	}

	map2["k1"] = "foo"

	eq = reflect.DeepEqual(map1, map2)
	if eq {
		t.Error("Maps are equal after change.")
	}
}

func TestPrintMapAsString(t *testing.T) {
	map1 := map[string]string{"k1": "v1", "k2": "v2"}
	result := PrintMapAsString(map1, ";", "=")
	expected := "k1=v1;k2=v2"
	if result != expected {
		t.Errorf("Failed: PrintMapAsString() expected:%s result: %s", expected, result)
	}

	map1 = map[string]string{"k1": "v1", "k2": "v2"}
	result = PrintMapAsString(map1, "", "")
	expected = "k1v1k2v2"
	if result != expected {
		t.Errorf("Failed: PrintMapAsString() expected:%s result: %s", expected, result)
	}

}
