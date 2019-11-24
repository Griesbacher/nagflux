package helper

import (
	"github.com/griesbacher/nagflux/config"
	"strings"
)

//SanitizeInfluxInput adds backslashes to special chars.
func SanitizeInfluxInput(input string) string {
	if len(input) == 0 {
		return input
	}
	if string(input[0]) == `"` && string(input[len(input)-1]) == `"` {
		return input
	}
	if config.GetConfig().InfluxDBGlobal.NastyString != "" {
		input = strings.Replace(
			input,
			config.GetConfig().InfluxDBGlobal.NastyString,
			config.GetConfig().InfluxDBGlobal.NastyStringToReplace,
			-1,
		)
	}
	input = strings.Trim(input, `'`)
	input = strings.Replace(input, " ", `\ `, -1)
	input = strings.Replace(input, ",", `\,`, -1)

	return input
}

//SanitizeInfluxField escapes '"' chars only
func SanitizeInfluxField(input string) string {

	if config.GetConfig().InfluxDBGlobal.NastyString != "" {
		input = strings.Replace(
			input,
			config.GetConfig().InfluxDBGlobal.NastyString,
			config.GetConfig().InfluxDBGlobal.NastyStringToReplace,
			-1,
		)
	}

	input = strings.Replace(input, "\"", `\"`, -1)

	return input
}

//SanitizeMap calls SanitizeInfluxInput in key and value
func SanitizeMap(input map[string]string) map[string]string {
	result := map[string]string{}
	for k, v := range input {
		result[SanitizeInfluxInput(k)] = SanitizeInfluxInput(v)
	}
	return result
}
