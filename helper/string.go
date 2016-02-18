package helper

import (
	"github.com/griesbacher/nagflux/logging"
	"strconv"
	"strings"
	"time"
)

//IsStringANumber returns true if the given string can be casted to int or float.
func IsStringANumber(input string) bool {
	_, floatErr := strconv.ParseFloat(input, 32)
	if floatErr == nil {
		return true
	}
	_, intErr := strconv.ParseInt(input, 10, 32)
	return intErr == nil
}

//StringToMap splits a string by two splitter an returns a map.
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

//StringIntToStringFloat adds a '.0' to a string if it does not contain a dot.
func StringIntToStringFloat(inputInt string) string {
	if inputInt == "" {
		return inputInt
	}

	if !strings.Contains(inputInt, ".") {
		inputInt += ".0"
	}
	return inputInt
}

//CastStringTimeFromSToMs adds three zeros to the timestring to cast from Seconds to Milliseconds.
func CastStringTimeFromSToMs(time string) string {
	return time + "000"
}

//GetYearMonthFromStringTimeMs returns the year and the month of a string which is in ms.
func GetYearMonthFromStringTimeMs(timeString string) (int, int) {
	i, err := strconv.ParseInt(timeString[:len(timeString)-3], 10, 64)
	if err != nil {
		logging.GetLogger().Warn(err.Error())
	}
	date := time.Unix(i, 0)
	return date.Year(), int(date.Month())
}
