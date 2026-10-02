package requests

type UpdateUserProfileRequest struct {
	Username string `json:"username" validate:"omitempty,min=3,max=30"`
}
