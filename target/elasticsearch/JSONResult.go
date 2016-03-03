package elasticsearch

//JSONResult is the JSON object returned from an bulk request
type JSONResult struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Create struct {
			ID    string `json:"_id"`
			Index string `json:"_index"`
			Type  string `json:"_type"`
			Error struct {
				CausedBy struct {
					Reason string `json:"reason"`
					Type   string `json:"type"`
				} `json:"caused_by"`
				Reason string `json:"reason"`
				Type   string `json:"type"`
			} `json:"error"`
			Shards struct {
				Failed     int `json:"failed"`
				Successful int `json:"successful"`
				Total      int `json:"total"`
			} `json:"_shards"`
			Version int `json:"_version"`
			Status  int `json:"status"`
		} `json:"create"`
	} `json:"items"`
	Took int `json:"took"`
}
