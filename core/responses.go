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
	ID         int    `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"userId"`
	Name       string `gorm:"column:name" json:"name"`
	SpriteType string `gorm:"column:spriteType" json:"spriteType"`
}

type getUserMapResponse struct {
	UserMap []userData `json:"userMap"`
	Success bool       `json:"success"`
}
