package helper

//SumIntSliceTillPos summarise an int slice till a given position.
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

//Contains checks if all values are within the list
func Contains(hay []string, needles []string) bool {
	hit := 0
	for _, a := range hay {
		for _, b := range needles {
			if a == b {
				hit++
			}
		}
	}
	return hit == len(needles)
}
