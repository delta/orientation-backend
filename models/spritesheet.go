package models

type SpriteSheet struct {
	ID int `gorm:"column:id;primary_key"`
	// add other neccessary fields
}

func (SpriteSheet) TableName() string {
	return "SpriteSheet"
}
