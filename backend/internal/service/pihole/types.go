package pihole

type TestExistingConnectionParams struct {
	Scheme   *string `json:"scheme"`
	Host     *string `json:"host"`
	Port     *int    `json:"port"`
	Password *string `json:"password"`
}

type TestInstanceConnectionParams struct {
	Scheme   string `json:"scheme"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
}
