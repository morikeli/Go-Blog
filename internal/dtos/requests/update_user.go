package requests

type UpdateUserProfileRequest struct {
	Username string `json:"username" form:"username" validate:"omitempty,min=3,max=30"`
}
