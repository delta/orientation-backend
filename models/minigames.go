package models

type MiniGame struct {
	ID   int    `gorm:"column:id;primaryKey;AUTO_INCREMENT"`
	Name string `gorm:"column:name;not null"`
}

func (MiniGame) TableName() string {
	return "MiniGame"
}
