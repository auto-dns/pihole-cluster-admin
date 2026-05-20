package piholesvc

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

type TestExistingConnectionCommand struct {
	Scheme   *string
	Host     *string
	Port     *int
	Password *string
}

type TestInstanceConnectionCommand struct {
	Scheme   string
	Host     string
	Port     int
	Password string
}

type RotatePasswordCommand struct {
	NewPassword string
}
