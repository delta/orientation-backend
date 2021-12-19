package models

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/delta/orientation-backend/config"
)

func CreateNewUser(email string, name string, gender Gender) User {
	rand.Seed(time.Now().UnixNano())
	user := User{Email: email, Name: name, Gender: gender}
	return user
}

// find the user with the given condition
// return the found user, and true if there was an error
func GetOnCondition(condition string, value interface{}) (User, bool) {
	cond := fmt.Sprintf("%s = ?", condition)
	var user User
	err := config.DB.Where(cond, value).First(&user).Error
	fmt.Println(err)
	if err != nil {
		return User{}, true
	} else {
		return user, false
	}
}

func Update(u *User) error {

	return config.DB.Model(u).Updates(u).Error
}
