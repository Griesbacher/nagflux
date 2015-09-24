package helper

import (
	"fmt"
	"strings"
)

//Creates a real copy of a string to string map.
func CopyMap(old map[string]string) map[string]string {
	newMap := map[string]string{}
	for k, v := range old {
		newMap[k] = v
	}
	return newMap
}

//Prints a map in the influxdb tags format.
func PrintMapAsString(toPrint map[string]string, fieldSeparator, assignmentSeparator string) string {
	return strings.Replace(strings.Replace(strings.Trim(fmt.Sprintf("%s", toPrint), "map[]"), " ", fieldSeparator, -1), ":", assignmentSeparator, -1)
}
