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
	ID           int    `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	Email        string `gorm:"column:email;unique;not null"`
	Name         string `gorm:"column:name"`
	Username     string `gorm:"column:userName;default:null"`
	Description  string `gorm:"column:description;default:null"`
	Gender       Gender `gorm:"column:gender"`
	Department   string `gorm:"column:department;default:null"`
	RefreshToken string `gorm:"column:refreshToken;default:null"`
	SpriteType   string `gorm:"column:spriteType"`
}

func (User) TableName() string {
	return "User"
}
