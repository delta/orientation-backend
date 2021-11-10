package models

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/delta/orientation-backend/config"
)

var totalSprites = 4

func CreateNewUser(email string, name string, gender Gender) User {
	rand.Seed(time.Now().UnixNano())
	user := User{Email: email, Name: name, Gender: gender, SpriteSheetID: rand.Intn(totalSprites) + 1}
	return user
}

func GetOnCondition(condition string, value string) (User, bool) {
	cond := fmt.Sprintf("%s = ?", condition)
	var user User
	err := config.DB.Where(cond, value).First(&user)
	if err != nil {
		return User{}, true
	} else {
		return user, false
	}
}
