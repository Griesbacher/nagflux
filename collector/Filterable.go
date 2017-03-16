package collector

import "strings"

//Filterable allows to sort the data
type Filterable struct {
	Filter string
}

//AllFilterable will be used by everybody
var AllFilterable = Filterable{Filter: All}

//EmptyFilterable is the default value
var EmptyFilterable = Filterable{Filter: ""}

//All will be used by everybody
const All = "all"

//TestTargetFilter tests if the given filter matches with the containing filter
func (f Filterable) TestTargetFilter(toTest string) bool {
	//temporary change the value to lower
	toTest = strings.ToLower(toTest)
	f.Filter = strings.ToLower(f.Filter)
	if f.Filter == toTest {
		return true
	}
	if f.Filter == All || toTest == All {
		return true
	}
	//Test Lists
	splitSource := strings.Split(f.Filter, ",")
	splitTarget := strings.Split(toTest, ",")
	for _, s := range splitSource {
		for _, t := range splitTarget {
			if s == t {
				return true
			}
		}
	}
	return false
}

//TestTargetFilterObj like TestTargetFilter just with two objects
func (f Filterable) TestTargetFilterObj(filter Filterable) bool {
	return filter.TestTargetFilter(f.Filter)
}
