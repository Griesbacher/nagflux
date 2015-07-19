package helper

import (
	"strings"
)

func StringToMap(input, entrySplitter, keyValueSplitter string) map[string]string {
	result := make(map[string]string)
	entry := strings.SplitAfter(input, entrySplitter)
	for _, pair := range entry {
		keyValue := strings.Split(strings.TrimSpace(pair), keyValueSplitter)
		result[keyValue[0]] = strings.Join(keyValue[1:], keyValueSplitter)
	}
	return result
}

func StringIntToStringFloat(inputInt string) string {
	if !strings.Contains(inputInt, ".") {
		inputInt += ".0"
	}
	return inputInt
}
