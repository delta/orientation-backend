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
)

func Auth(c echo.Context) error {
	userNewCookie, isNewCookie, isLoggedIn := checkAuth(c)
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
		"client_id":     currentConfig.Dauth.Client_id,
		"redirect_uri":  config.Config("CALLBACK_PAGE_URI"),
		"response_type": currentConfig.Dauth.Response_type,
		"grant_type":    currentConfig.Dauth.Grant_type,
		"state":         currentConfig.Dauth.State,
		"scope":         strings.Join(currentConfig.Dauth.Scope, "+"),
		"nonce":         currentConfig.Dauth.Nonce,
	}
	queryString, err := encodeQuery(params)
	if err != nil {
		panic(fmt.Errorf("Error with url"))
	}
	base, _ := url.Parse("https://auth.delta.nitt.edu/authorize")
	base.RawQuery = queryString.Encode()
	return c.Redirect(http.StatusFound, base.String())
}

func CallBack(c echo.Context) error {
	code := c.QueryParam("code")
	fmt.Println(code)

	// Getting Token
	userToken, refreshToken, _, err := handleCallBack(code)
	if err != nil {
		return c.JSON(http.StatusOK, isAuthResult{Status: false})
	}

	userCookie := http.Cookie{
		Name:     currentConfig.Cookie.User.Name,
		Value:    userToken,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((time.Duration(currentConfig.Cookie.User.Expires) * time.Hour).Seconds()),
	}

	refreshCookie := http.Cookie{
		Name:     currentConfig.Cookie.Refresh.Name,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((time.Duration(currentConfig.Cookie.Refresh.Expires) * time.Hour).Seconds()),
	}
	c.SetCookie(&userCookie)
	c.SetCookie(&refreshCookie)
	origin := c.Request().Header.Get(echo.HeaderOrigin)
	c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, origin)
	c.Response().Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
	return c.JSON(http.StatusOK, isAuthResult{Status: true})
}

func LogOut(c echo.Context) error {
	fmt.Println("Here")
	userCookie := http.Cookie{
		Name:     currentConfig.Cookie.User.Name,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	c.SetCookie(&userCookie)
	refreshCookie := http.Cookie{
		Name:     currentConfig.Cookie.Refresh.Name,
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

func CheckAuth(c echo.Context) error {
	fmt.Println("Entered")
	newUserCookie, setCookie, result := checkAuth(c)
	if setCookie {
		c.SetCookie(&newUserCookie)
	}
	origin := c.Request().Header.Get(echo.HeaderOrigin)
	c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, origin)
	c.Response().Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
	return c.JSON(http.StatusOK, isAuthResult{Status: result})
}
