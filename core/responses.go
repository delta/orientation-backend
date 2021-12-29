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

type userData struct {
	UserId     int    `json:"id"`
	Name       string `json:"name"`
	SpriteType string `json:"spriteType"`
}

type getUserMapResponse struct {
	UserMap []userData `json:"userMap"`
	Success bool       `json:"success"`
}
