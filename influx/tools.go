package influx
import "strings"

func SanitizeInput(input string) string {
	input = strings.Replace(input, "\\", "\\\\", -1)
	input = strings.Replace(input, " ", "\\ ", -1)
	input = strings.Replace(input, ",", "\\,", -1)
	return input
}