package helper

import "strings"

//Adds backslashes to special chars.
func SanitizeInfluxInput(input string) string {
	input = strings.Replace(input, "\\", "\\\\", -1)
	input = strings.Replace(input, " ", "\\ ", -1)
	input = strings.Replace(input, ",", "\\,", -1)
	return input
}
