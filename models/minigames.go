package models

type MiniGame struct {
	ID     int    `gorm:"column:id;primaryKey;AUTO_INCREMENT"`
	Name   string `gorm:"column:name;not null"`
	RoomID int    `gorm:"column:roomId; not null"`
}

func (MiniGame) TableName() string {
	return "MiniGame"
}
