package elasticsearch

//JSONResult is the JSON object returned from an bulk request
type JSONResult struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Create struct {
			_id    string `json:"_id"`
			_index string `json:"_index"`
			_type  string `json:"_type"`
			Error  struct {
				CausedBy struct {
					Reason string `json:"reason"`
					Type   string `json:"type"`
				} `json:"caused_by"`
				Reason string `json:"reason"`
				Type   string `json:"type"`
			} `json:"error"`
			_shards struct {
				Failed     int `json:"failed"`
				Successful int `json:"successful"`
				Total      int `json:"total"`
			} `json:"_shards"`
			_version int `json:"_version"`
			Status   int `json:"status"`
		} `json:"create"`
	} `json:"items"`
	Took int `json:"took"`
}
