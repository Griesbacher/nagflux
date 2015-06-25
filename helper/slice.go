package helper

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
