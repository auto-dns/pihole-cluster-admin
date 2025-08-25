package cluster

type getSummaryDTO struct {
	Mode      string `json:"mode"`
	Unanimous bool   `json:"unanimous"`
	Counts    struct {
		Total    int `json:"total"`
		Enabled  int `json:"enabled"`
		Disabled int `json:"disabled"`
		Failed   int `json:"failed"`
		Errors   int `json:"errors"`
	} `json:"counts"`
	Timers struct {
		Present    bool   `json:"present"`
		MinSeconds *int64 `json:"minSeconds,omitempty"`
		MaxSeconds *int64 `json:"maxSeconds,omitempty"`
	} `json:"timers"`
	Took struct {
		MaxSeconds float64 `json:"maxSeconds"`
		AvgSeconds float64 `json:"avgSeconds"`
	} `json:"took"`
}

type getNodeDTO struct {
	Node struct {
		Id   int64  `json:"id"`
		Name string `json:"name"`
		Host string `json:"host"`
	} `json:"node"`
	Blocking string  `json:"blocking"`
	Timer    *int64  `json:"timer,omitempty"`
	Took     float64 `json:"took"`
	Error    string  `json:"error,omitempty"`
}

type getResponseDTO struct {
	Summary getSummaryDTO        `json:"summary"`
	Nodes   map[int64]getNodeDTO `json:"nodes"`
}
