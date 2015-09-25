package helper

import (
	"strings"
)

//Splits a string by two splitter an returns a map.
func StringToMap(input, entrySplitter, keyValueSplitter string) map[string]string {
	if entrySplitter == "" || keyValueSplitter == "" || input == "" {
		return nil
	}

	result := make(map[string]string)
	entry := strings.Split(input, entrySplitter)
	for _, pair := range entry {
		keyValue := strings.Split(strings.TrimSpace(pair), keyValueSplitter)
		result[keyValue[0]] = strings.Join(keyValue[1:], keyValueSplitter)
	}
	return result
}

//Adds a '.0' to a string if it does not contain a dot.
func StringIntToStringFloat(inputInt string) string {
	if inputInt == "" {
		return inputInt
	}

	if !strings.Contains(inputInt, ".") {
		inputInt += ".0"
	}
	return inputInt
}

//Adds three zeros to the timestring to cast from Seconds to Milliseconds.
func CastStringTimeFromSToMs(time string) string {
	return time + "000"
}
