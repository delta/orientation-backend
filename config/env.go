package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config func to get env
func Config(key string) string {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		panic(fmt.Errorf("error loading .env file, %+v", err))
	}
	return os.Getenv(key)
}
