package core

import (
	"net/http"

	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/models"
	"github.com/labstack/echo/v4"
	logger "github.com/sirupsen/logrus"
)

func RegisterRoutes(v *echo.Group) {
	v.GET("/user/me", GetUserData)
	v.PUT("/user/signup", UpdateUserData)
	v.GET("/user/map", getUserMap)
}

// returns the user data with the given credentials
func GetUserData(c echo.Context) error {
	l := logger.WithFields(logger.Fields{"method": "core/getUserData"})

	l.Infof("user has requested for user data")

	l.Debugf("Trying to fetch user Data from cookie")
	user, err := GetCurrentUser(c)

	l.Debugf("Found the user=%v while requesting for user", user)

	if err != nil {
		l.Errorf("Couldn't find user, because %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User not authenticated"})
	}

	l.Infof("Successfully found the user")

	return c.JSON(http.StatusOK, getUserDataResponse{User: user, IsNewUser: user.Description == ""})
}

func UpdateUserData(c echo.Context) error {
	l := logger.WithFields(logger.Fields{"method": "core/updateUserData"})
	l.Infof("user has requested for to update User data")

	l.Debugf("Trying to fetch user Data from cookie")
	u, err := GetCurrentUser(c)

	l.Debugf("Found the user=%v while requesting for user", u)
	if err != nil {
		l.Errorf("Couldn't find user, because %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User not authenticated"})
	}

	l.Debugf("Successfully Found the user, validating the user data")

	req := newUserUpdateRequest()
	req.populate(&u)

	if err := req.bind(c, &u); err != nil {
		l.Errorf("Couldn't validate user data, because %v", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Bad Request", Error: err})
	}

	l.Debugf("Successfully validated user data, trying to save the user data")
	if err := models.Update(&u); err != nil {
		l.Errorf("Couldn't save user data, because %v", err)
		return c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Message: "Internal Server Error", Error: err})
	}

	l.Debugf("Successfully saved the user data")

	l.Infof("Successfully updated user data")

	return c.JSON(http.StatusOK, updateUserDataResponse{Success: true, User: u})
}

func getUserMap(c echo.Context) error {
	l := logger.WithFields(logger.Fields{"method": "core/getUserMap"})

	u, err := GetCurrentUser(c)

	l.Debugf("Found the user=%v while requesting for user", u)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User not authenticated"})
	}

	db := config.DB

	var userMap []userData

	if err := db.Select("id", "name", "spriteType").Find(&userMap).Error; err != nil {
		l.Errorf("Erorr %e fetching user map from db", err)
		return c.JSON(http.StatusInternalServerError, getUserMapResponse{UserMap: userMap, Success: false})
	}

	return c.JSON(http.StatusAccepted, getUserMapResponse{UserMap: userMap, Success: true})
}
