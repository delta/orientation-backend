package models

import "database/sql/driver"

type Gender string

const (
	Male   Gender = "male"
	Female Gender = "female"
)

func (e *Gender) Scan(value interface{}) error {
	*e = Gender(value.([]byte))
	return nil
}

func (e Gender) Value() (driver.Value, error) {
	return string(e), nil
}

type User struct {
	ID            int    `gorm:"primary_key;auto_increment" json:"-"`
	Email         string `gorm:"unique;not null"`
	Name          string
	Description   string `gorm:"default:null"`
	Gender        Gender `sql:"type:ENUM('male', 'female')"`
	Department    string `gorm:"default:null"`
	RefreshToken  string `gorm:"default:null"`
	SpriteSheetID int
	Spritesheet   SpriteSheet `gorm:"foreignKey:SpriteSheetID"`
}
