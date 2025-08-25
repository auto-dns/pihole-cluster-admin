package pihole

// General

// Blocking

type blockingWireResponse struct {
	Blocking string  `json:"blocking"` // "enabled"|"disabled"|"failed"|"unknown"
	Timer    *int64  `json:"timer"`
	Took     float64 `json:"took"`
}
