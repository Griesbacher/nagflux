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
	result := make(map[string]string)
	if entrySplitter == "" || keyValueSplitter == "" || input == "" {
		return result
	}

	entry := strings.Split(input, entrySplitter)
	for _, pair := range entry {
		keyValue := strings.Split(strings.TrimSpace(pair), keyValueSplitter)
		value := strings.Join(keyValue[1:], keyValueSplitter)
		if value != "" && keyValue[0] != "" {
			result[keyValue[0]] = value
		}
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

//VersionOrdinal from here: https://stackoverflow.com/questions/18409373/how-to-compare-two-version-number-strings-in-golang/18411978#18411978
func VersionOrdinal(version string) string {
	// ISO/IEC 14651:2011
	const maxByte = 1<<8 - 1
	vo := make([]byte, 0, len(version)+8)
	j := -1
	for i := 0; i < len(version); i++ {
		b := version[i]
		if '0' > b || b > '9' {
			vo = append(vo, b)
			j = -1
			continue
		}
		if j == -1 {
			vo = append(vo, 0x00)
			j = len(vo) - 1
		}
		if vo[j] == 1 && vo[j+1] == '0' {
			vo[j+1] = b
			continue
		}
		if vo[j]+1 > maxByte {
			panic("VersionOrdinal: invalid version")
		}
		vo = append(vo, b)
		vo[j]++
	}
	return string(vo)
}
