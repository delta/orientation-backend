package core

import "github.com/delta/orientation-backend/models"

type getUserDataResponse struct {
	User      models.User `json:"user"`
	IsNewUser bool        `json:"isNewUser"`
}

type updateUserDataResponse struct {
	User    models.User `json:"user"`
	Success bool        `json:"success"`
}
