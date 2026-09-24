package responses

import (
	"github.com/morikeli/golangrestapi/internal/store"
)

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	User         UserResponse `json:"user"`
}
