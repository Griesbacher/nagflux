package helper

import "testing"

var SumIntSliceTillPosData = []struct {
	slice  []int
	pos    int
	output int
}{
	{[]int{1, 1, 1, 1}, 2, 3},  //simple add
	{[]int{0, 0, 0, 0}, 3, 0},  //just zeros
	{[]int{1, 1, 1, 1}, 10, 4}, //out of range -> take whole array
	{[]int{1, 1, 1, 1}, -1, 0}, //out of range -> take nothing
}

func TestSumIntSliceTillPos(t *testing.T) {
	for _, data := range SumIntSliceTillPosData {
		actual := SumIntSliceTillPos(data.slice, data.pos)
		if actual != data.output {
			t.Errorf("SanitizeInfluxData(%d): expected:%d, actual:%d", data.slice, data.output, actual)
		}
	}
}
