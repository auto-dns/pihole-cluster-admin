package user

type PatchUserCommand struct {
	Username *string
}

type UpdatePasswordCommand struct {
	CurrentPassword string
	NewPassword     string
}
