package config

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB connection object, this will be used to connect and do operation in db
var DB *gorm.DB

func initDB() {
	dbName := Config("DB_NAME")
	dbPwd := Config("DB_PWD")
	dbUser := Config("DB_USER")
	dbHost := Config("DB_HOST")
	dbPort := Config("DB_PORT")

	// db connection str
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true", dbUser, dbPwd, dbHost, dbPort, dbName)

	//connecting to db
	db, err := gorm.Open(mysql.Open(connStr), &gorm.Config{})

	if err != nil {
		panic(fmt.Errorf("error connecting DB, %+v", err))
	}

	DB = db
}
