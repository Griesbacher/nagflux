package helper

import (
	"fmt"
	"strings"
)

//CreateJSONFromStringMap creates a part of a JSON object
func CreateJSONFromStringMap(input map[string]string) string {
	result := ""
	for k, v := range input {
		result += fmt.Sprintf(`,"%s":%s`, k, GenJSONValueString(v))
	}
	return result
}

//GenJSONValueString quotes the string if it's not a number.
func GenJSONValueString(input string) string {
	if IsStringANumber(input) {
		return input
	}
	return fmt.Sprintf(`"%s"`, input)
}

//SanitizeElasicInput escapes backslashes and trims single ticks.
func SanitizeElasicInput(input string) string {
	input = strings.Trim(input, `'`)
	input = strings.Replace(input, `\`, `\\`, -1)
	input = strings.Replace(input, `"`, `\"`, -1)
	return input
}
