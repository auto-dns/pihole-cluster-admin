package user

type PatchUserParams struct {
	Username *string `json:"username"`
}

type UpdatePasswordParams struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}
