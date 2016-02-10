package helper

import "fmt"

//CreateJSONFromStringMap creates a part of a JSON object
func CreateJSONFromStringMap(input map[string]string) string {
	result := ""
	for k, v := range input {
		result += fmt.Sprintf(`,"%s":`, k)
		if IsStringANumber(v) {
			result += v
		} else {
			result += fmt.Sprintf(`"%s"`, v)
		}
	}
	return result
}
