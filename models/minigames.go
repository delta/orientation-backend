package models

import "time"

type MiniGames struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"unique"`
	RoomID    int
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}
