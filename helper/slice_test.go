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
var ContainsData = []struct {
	hay    []string
	needls []string
	output bool
}{
	{[]string{"a", "b", "c"}, []string{"c", "b", "a"}, true},
	{[]string{"a", "b", "c"}, []string{"b", "a"}, true},
	{[]string{"a", "b", "c"}, []string{}, true},
	{[]string{"a", "b", "c"}, []string{"x"}, false},
	{[]string{"a", "b"}, []string{"c", "b", "a"}, false},
}

func TestSumIntSliceTillPos(t *testing.T) {
	t.Parallel()
	for _, data := range SumIntSliceTillPosData {
		actual := SumIntSliceTillPos(data.slice, data.pos)
		if actual != data.output {
			t.Errorf("SanitizeInfluxData(%d): expected:%d, actual:%d", data.slice, data.output, actual)
		}
	}
}

func TestContains(t *testing.T) {
	t.Parallel()
	for _, data := range ContainsData {
		actual := Contains(data.hay, data.needls)
		if actual != data.output {
			t.Errorf("SanitizeInfluxData(%s): expected:%t, actual:%t", data.hay, data.output, actual)
		}
	}
}
