package auth

import "github.com/delta/orientation-backend/models"

type CallbackResponse struct {
	User       models.User  `json:"user"`
	AuthStatus isAuthResult `json:"authStatus"`
}

type TokenResult struct {
	Type    string `json:"token_type"`
	Token   string `json:"access_token"`
	State   string `json:"state"`
	Expires int64  `json:"expires_in"`
}

type UserResult struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	Gender string `json:"gender"`
}

type isAuthResult struct {
	Status    bool `json:"status"`
	IsNewUser bool `json:"isNewUser"`
}
