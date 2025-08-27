package pihole

type AddNodeCommand struct {
	Scheme      string
	Host        string
	Port        int
	Name        string
	Description string
	Password    string
}

type UpdateNodeCommand struct {
	Scheme      *string
	Host        *string
	Port        *int
	Name        *string
	Description *string
	Password    *string
}

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
