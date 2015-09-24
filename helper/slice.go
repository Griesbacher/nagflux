package helper

//Summarise an int slice till a given position.
func SumIntSliceTillPos(slice []int, pos int) int {
	sum := 0
	for index, value := range slice {
		if index <= pos {
			sum += value
		} else {
			break
		}
	}
	return sum
}

//Loops of slice and checks if it contains a given string.
func SliceContainsString(str string, slice []string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

//Removes duplicates strings from a string slice - UNUSED.
func RemoveDuplicateStrings(dupes []string) []string {
	emptySlice := []string{}
	for _, value := range dupes {
		if !SliceContainsString(value, emptySlice) {
			emptySlice = append(emptySlice, value)
		}
	}
	return emptySlice
}
