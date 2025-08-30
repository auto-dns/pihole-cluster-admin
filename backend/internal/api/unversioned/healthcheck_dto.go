package unversioned

import "encoding/json"

type HealthcheckStatus string

const (
	HealthcheckStatusOk      HealthcheckStatus = "ok"
	HealthcheckStatusUnready HealthcheckStatus = "unready"
)

func (hs HealthcheckStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(hs))
}

type healthcheckResponseDTO struct {
	Status HealthcheckStatus `json:"status"`
}
