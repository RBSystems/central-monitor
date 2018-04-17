package elk

type BaseResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:":timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total int       `json:"total"`
		Max   int       `json:"max_store"`
		Hits  []BaseHit `json:"hits"`
	} `json:"hits"`
}

type BaseHit struct {
	Index string  `json:"_index"`
	Type  string  `json:"_type"`
	ID    string  `json:"_id"`
	Score float64 `json:"_score"`
}
