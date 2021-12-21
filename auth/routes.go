package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	// "github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/config"
	"github.com/labstack/echo/v4"
	logger "github.com/sirupsen/logrus"
	"strconv"
)

// Route to register all auth routes
func RegisterRoutes(v *echo.Group) {
	v.GET("/callback", CallBack)
	v.GET("/logout", LogOut)
	v.GET("/checkAuth", CheckAuth_deprecated)
	if is_dev {
		v.POST("/dummyLogin", DummyLogin)
	}
}

func Auth(c echo.Context) error {
	userNewCookie, isNewCookie, isLoggedIn := CheckAuth(c)
	if isLoggedIn {
		if isNewCookie {
			c.SetCookie(&userNewCookie)
		}
		return c.Redirect(http.StatusFound, uiURL)
	}
	c.Response().Header().Set("Content-Type", "application/json")
	fmt.Println("Inside new auth")
	// fmt.Println(r.Host)
	params := map[string]string{
		"client_id":     CurrentConfig.Dauth.Client_id,
		"redirect_uri":  config.Config("CALLBACK_PAGE_URI"),
		"response_type": CurrentConfig.Dauth.Response_type,
		"grant_type":    CurrentConfig.Dauth.Grant_type,
		"state":         CurrentConfig.Dauth.State,
		"scope":         strings.Join(CurrentConfig.Dauth.Scope, "+"),
		"nonce":         CurrentConfig.Dauth.Nonce,
	}
	queryString, err := encodeQuery(params)
	if err != nil {
		panic(fmt.Errorf("error with url"))
	}
	base, _ := url.Parse("https://auth.delta.nitt.edu/authorize")
	base.RawQuery = queryString.Encode()
	return c.Redirect(http.StatusFound, base.String())
}

func DummyLogin(c echo.Context) error {
	type UserStruct struct {
		Roll string `json:"roll"`
	}
	u := new(UserStruct)
	c.Bind(u)
	fmt.Println(u.Roll)
	r, err := strconv.Atoi(u.Roll)
	if err != nil {
		return c.JSON(http.StatusOK, isAuthResult{Status: false})
	}
	userToken, refreshToken := createDummyUser(r)
	userCookie := http.Cookie{
		Name:     CurrentConfig.Cookie.User.Name,
		Value:    userToken,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((time.Duration(CurrentConfig.Cookie.User.Expires) * time.Hour).Seconds()),
	}

	refreshCookie := http.Cookie{
		Name:     CurrentConfig.Cookie.Refresh.Name,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((time.Duration(CurrentConfig.Cookie.Refresh.Expires) * time.Hour).Seconds()),
	}
	c.SetCookie(&userCookie)
	c.SetCookie(&refreshCookie)
	origin := c.Request().Header.Get(echo.HeaderOrigin)
	c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, origin)
	c.Response().Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
	return c.JSON(http.StatusOK, isAuthResult{Status: true})
}

// Handles dauth callback
// and returns if the authentication was successful
func CallBack(c echo.Context) error {
	code := c.QueryParam("code")
	fmt.Println(code)

	// Getting Token
	userToken, refreshToken, user, isNewUser, err := handleCallBack(code)
	if err != nil {
		return c.JSON(http.StatusOK, isAuthResult{Status: false})
	}

	userCookie := http.Cookie{
		Name:     CurrentConfig.Cookie.User.Name,
		Value:    userToken,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((time.Duration(CurrentConfig.Cookie.User.Expires) * time.Hour).Seconds()),
	}

	refreshCookie := http.Cookie{
		Name:     CurrentConfig.Cookie.Refresh.Name,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((time.Duration(CurrentConfig.Cookie.Refresh.Expires) * time.Hour).Seconds()),
	}
	c.SetCookie(&userCookie)
	c.SetCookie(&refreshCookie)
	origin := c.Request().Header.Get(echo.HeaderOrigin)
	c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, origin)
	c.Response().Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
	return c.JSON(http.StatusOK, CallbackResponse{User: user, AuthStatus: isAuthResult{Status: true, IsNewUser: isNewUser}})
}

func LogOut(c echo.Context) error {
	// fmt.Println("Here")
	userCookie := http.Cookie{
		Name:     CurrentConfig.Cookie.User.Name,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	c.SetCookie(&userCookie)
	refreshCookie := http.Cookie{
		Name:     CurrentConfig.Cookie.Refresh.Name,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	c.SetCookie(&refreshCookie)
	origin := c.Request().Header.Get(echo.HeaderOrigin)
	c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, origin)
	c.Response().Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
	return c.JSON(http.StatusOK, isAuthResult{Status: true})
}

func CheckAuth_deprecated(c echo.Context) error {
	var l = logger.WithFields(logger.Fields{
		"method": "Auth/Checkauth",
	})
	l.Debugf("Entered")
	newUserCookie, setCookie, result := CheckAuth(c)
	if setCookie {
		c.SetCookie(&newUserCookie)
	}
	origin := c.Request().Header.Get(echo.HeaderOrigin)
	c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, origin)
	c.Response().Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
	return c.JSON(http.StatusOK, isAuthResult{Status: result})
}
