package responses

import (
	"github.com/morikeli/golangrestapi/internal/store"
)

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	User         *store.User  `json:"user"`
}
