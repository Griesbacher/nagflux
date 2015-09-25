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
