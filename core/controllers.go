package core

import (
	"fmt"

	"github.com/delta/orientation-backend/auth"
	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/models"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	// logger "github.com/sirupsen/logrus"
)

// Gets the currently logged in user with the cookie
func GetCurrentUser(c echo.Context) (models.User, error) {
	cookie, err := c.Cookie(auth.CurrentConfig.Cookie.User.Name)
	if err != nil {
		// fmt.Println("No cookie")
		return models.User{}, fmt.Errorf("couldn't find cookie")
	}
	token, err := jwt.ParseWithClaims(cookie.Value, &auth.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return auth.HmacSampleSecret, nil
	})
	if err != nil {
		return models.User{}, err
	}
	if claims, ok := token.Claims.(*auth.CustomClaims); ok && token.Valid {
		var user models.User
		err = nil
		err = config.DB.Where("email = ?", claims.Email).First(&user).Error
		return user, err
	} else {
		return models.User{}, fmt.Errorf("invalid token")
	}
}
