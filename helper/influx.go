package helper

import "strings"

//SanitizeInfluxInput adds backslashes to special chars.
func SanitizeInfluxInput(input string) string {
	input = strings.Trim(input, `'`)
	input = strings.Replace(input, `\`, `\\`, -1)
	input = strings.Replace(input, " ", `\ `, -1)
	input = strings.Replace(input, ",", `\,`, -1)
	return input
}
