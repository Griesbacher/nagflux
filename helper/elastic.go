package helper

import (
	"fmt"
	"github.com/griesbacher/nagflux/config"
	"strings"
)

//CreateJSONFromStringMap creates a part of a JSON object
func CreateJSONFromStringMap(input map[string]string) string {
	result := ""
	for k, v := range input {
		result += fmt.Sprintf(`,%s:%s`, GenJSONValueString(k), GenJSONValueString(v))
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

//GenIndex generates an index depending on the config, ending with year and month
func GenIndex(index, timeString string) string {
	rotation := config.GetConfig().Elasticsearch.IndexRotation
	year, month := GetYearMonthFromStringTimeMs(timeString)
	switch rotation {
	case "monthly":
		return fmt.Sprintf("%s-%d.%02d", index, year, month)
	case "yearly":
		return fmt.Sprintf("%s-%d", index, year)
	default:
		panic(fmt.Sprintf("The given IndexRotation[%s] is not supported", rotation))
	}
}
