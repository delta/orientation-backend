package models

import (
	"time"
)

type Room struct {
	ID        int       `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"unique"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func GetAllRooms() ([]string, error) {
	var rooms []string

	if err := db.Model(&Room{}).Select("name").Find(&rooms).Error; err != nil {
		return nil, err
	}

	return rooms, nil
}
