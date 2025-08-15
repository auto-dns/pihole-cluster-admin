package pihole

// Pihole Auth DTO

type authResponse struct {
	Session struct {
		Valid    bool   `json:"valid"`
		SID      string `json:"sid"`
		CSRF     string `json:"csrf"`
		Validity int    `json:"validity"`
	} `json:"session"`
	Took float64 `json:"took"`
}
