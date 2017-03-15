package collector

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
	if f.Filter == All {
		return true
	}
	return f.Filter == toTest
}

//TestTargetFilterObj like TestTargetFilter just with two objects
func (f Filterable) TestTargetFilterObj(filter Filterable) bool {
	return filter.TestTargetFilter(f.Filter)
}
