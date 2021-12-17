package models

import (
	"github.com/delta/orientation-backend/config"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var db *gorm.DB
var log *logrus.Logger

// initilalizes config in models and run migrations
func Init() {
	db = config.DB
	log = config.Log
	// run migrations
	db.AutoMigrate(&User{}, &SpriteSheet{}, &Room{})
}
