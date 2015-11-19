package helper

import (
	"strings"
)

//CopyMap creates a real copy of a string to string map.
func CopyMap(old map[string]string) map[string]string {
	newMap := map[string]string{}
	for k, v := range old {
		newMap[k] = v
	}
	return newMap
}

//PrintMapAsString prints a map in the influxdb tags format.
func PrintMapAsString(toPrint map[string]string, fieldSeparator, assignmentSeparator string) string {
	result := ""
	for key, value := range toPrint {
		result += key + assignmentSeparator + value + fieldSeparator
	}
	result = strings.Trim(result, fieldSeparator)
	return result
}
