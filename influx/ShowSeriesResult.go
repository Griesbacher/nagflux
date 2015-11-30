package influx

//ShowSeriesResult represents the JSON query result
type ShowSeriesResult struct {
	Results ResultsStruct
}

//ResultsStruct is a list of series
type ResultsStruct []struct {
	Series SeriesStruct
}

//SeriesStruct is a list of field values
type SeriesStruct []struct {
	Columns []string
	Name    string
	Values  [][]string
}
