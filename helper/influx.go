package helper

import (
	"github.com/griesbacher/nagflux/config"
	"strings"
)

//SanitizeInfluxInput adds backslashes to special chars.
func SanitizeInfluxInput(input string) string {
	input = strings.Replace(input, config.GetConfig().Influx.NastyString, config.GetConfig().Influx.NastyStringToReplace, -1)
	input = strings.Trim(input, `'`)
	input = strings.Replace(input, " ", `\ `, -1)
	input = strings.Replace(input, ",", `\,`, -1)

	return input
}
