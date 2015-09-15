package helper

import (
	"fmt"
	"strings"
)

func CopyMap(old map[string]string) map[string]string {
	newMap := map[string]string{}
	for k, v := range old {
		newMap[k] = v
	}
	return newMap
}

func PrintMapAsString(toPrint map[string]string, fieldSeparator, assignmentSeparator string) string {
	return strings.Replace(strings.Replace(strings.Trim(fmt.Sprintf("%s", toPrint), "map[]"), " ", fieldSeparator, -1), ":", assignmentSeparator, -1)
}
