package v1

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
