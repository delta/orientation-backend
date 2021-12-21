package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/models"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	logger "github.com/sirupsen/logrus"
)

var authRoutes []string = []string{"/api/auth/callback", "/api/user/signup", "/api/auth/dummyLogin"}

func stringInSlice(s string, list []string) bool {
	for _, a := range list {
		if a == s {
			fmt.Println("Present")
			return true
		}
	}
	return false
}

func SkipperFunc(c echo.Context) bool {
	return (stringInSlice(c.Path(), authRoutes))

}

type (
	AuthMiddlewareConfig struct {
		Skipper Skipper
	}
	// A list of routes to skip authentication
	Skipper func(c echo.Context) bool
)

// Validates whether the user has logged in
func AuthMiddlewareWrapper(config AuthMiddlewareConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// skip authCheck for routes which are whilelisted by skipper
			if config.Skipper != nil && config.Skipper(c) {
				return next(c)
			}
			_, _, isLoggedIn := CheckAuth(c)
			if !isLoggedIn {
				return c.JSON(http.StatusForbidden, ErrorResponse{Message: "User not authenticated"})
			}
			return next(c)
		}
	}
}

// Verifies if the user has been authenticated
// returns httpUserCookie, isNewCookie, isLoggedIn
func CheckAuth(c echo.Context) (http.Cookie, bool, bool) {
	userCookie, err := c.Cookie(CurrentConfig.Cookie.User.Name)
	var l = logger.WithFields(logger.Fields{
		"method":     "auth/controllers/checkAuth",
		"userCookie": userCookie,
	})
	l.Infof("Trying to authenticate user")
	type userClaims struct {
		Email string `json:"email"`
		ID    int    `json:"id"`
		jwt.StandardClaims
	}
	if err == nil {
		token, err := jwt.ParseWithClaims(userCookie.Value, &userClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				errMsg := fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"])
				l.Errorf(errMsg)
				return nil, errors.New(errMsg)
			}

			return HmacSampleSecret, nil
		})
		if err == nil {
			if _, ok := token.Claims.(*userClaims); ok && token.Valid {
				l.Infof("user is logged in and has a valid jwt token")
				return http.Cookie{}, false, true
			}
		}
	}
	type refreshClaims struct {
		ID int `json:"id"`
		jwt.StandardClaims
	}
	l.Debugf("user's auth token has expired, now checking refresh token")
	refreshCookie, err := c.Cookie(CurrentConfig.Cookie.Refresh.Name)
	if err != nil {
		l.Infof("The user has not logged in")
		return http.Cookie{}, false, false
	}
	token, err := jwt.ParseWithClaims(refreshCookie.Value, &refreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			errMsg := fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"])
			l.Errorf(errMsg)
			return nil, errors.New(errMsg)
		}

		return HmacSampleSecret, nil
	})
	if err != nil {
		l.Errorf("Unable to verify user's refresh token due to %v", err)
		return http.Cookie{}, false, false
	}
	refresh, ok := token.Claims.(*refreshClaims)
	if !ok || !token.Valid {
		l.Errorf("couldn't verify user's refresh token")
		return http.Cookie{}, false, false
	}
	var user models.User
	err = config.DB.Where("id = ?", refresh.ID).First(&user).Error
	if err != nil {
		l.Errorf("Unable to fetch user's details from db")
		return http.Cookie{}, false, false
	}
	if token.Raw != user.RefreshToken {
		l.Errorf("Invalid refresh token")
		return http.Cookie{}, false, false
	}
	l.Debugf("User's jwt token has expired. Trying to create a refresh token")
	userToken, _ := createToken(jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(time.Duration(CurrentConfig.Cookie.User.Expires) * time.Hour).Unix(),
	})
	userNewCookie := http.Cookie{
		Name:   CurrentConfig.Cookie.User.Name,
		Value:  userToken,
		Path:   "/",
		MaxAge: int((time.Duration(CurrentConfig.Cookie.User.Expires) * time.Hour).Seconds()),
	}
	c.SetCookie(&userNewCookie)
	l.Infof("created a new http token for the user")
	return userNewCookie, true, true
}
