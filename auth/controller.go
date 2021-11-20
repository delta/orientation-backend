package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/models"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func handleCallBack(code string) (string, string, bool, error) {
	params := map[string]string{
		"code":          code,
		"client_secret": currentConfig.Dauth.Client_secret,
		"client_id":     currentConfig.Dauth.Client_id,
		"redirect_uri":  config.Config("CALLBACK_PAGE_URI"),
		"grant_type":    currentConfig.Dauth.Grant_type,
	}
	tokenQueryString, err := encodeQuery(params)
	if err != nil {
		panic(fmt.Errorf("Error with url"))
	}
	tokenEncodedData := tokenQueryString.Encode()
	tokenHeader := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	tokenResponse, err := makeRequest(http.MethodPost, "https://auth.delta.nitt.edu/api/oauth/token", tokenHeader, tokenEncodedData)
	if err != nil {
		return "", "", false, err
	}
	var tokenResult TokenResult
	err = json.Unmarshal(tokenResponse, &tokenResult)
	if err != nil {
		return "", "", false, err
	}
	resourceHeader := map[string]string{"Authorization": "Bearer " + tokenResult.Token}
	userResponse, err := makeRequest(http.MethodPost, "https://auth.delta.nitt.edu/api/resources/user", resourceHeader, url.Values{}.Encode())
	if err != nil {
		return "", "", false, err
	}
	var userResult UserResult
	err = json.Unmarshal(userResponse, &userResult)
	if err != nil {
		return "", "", false, err
	}
	var gender models.Gender
	if userResult.Gender == "MALE" {
		gender = models.Male
	} else {
		gender = models.Female
	}
	user, isUser := models.GetOnCondition("email", userResult.Email)
	isNewUser := false
	if isUser {
		isNewUser = true
		user = models.CreateNewUser(userResult.Email, userResult.Name, gender)
	}
	userToken, _ := createToken(jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(time.Duration(currentConfig.Cookie.User.Expires) * time.Hour).Unix(),
	})

	refreshToken, _ := createToken(jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().Add(time.Duration(currentConfig.Cookie.Refresh.Expires) * time.Hour).Unix(),
	})
	user.RefreshToken = refreshToken
	config.DB.Save(&user)
	return userToken, refreshToken, isNewUser, nil
}

func checkAuth(c echo.Context) (http.Cookie, bool, bool) {
	userCookie, err := c.Cookie(currentConfig.Cookie.User.Name)
	type userClaims struct {
		Email string `json:"email"`
		ID    int    `json:"id"`
		jwt.StandardClaims
	}
	if err == nil {
		token, err := jwt.ParseWithClaims(userCookie.Value, &userClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return hmacSampleSecret, nil
		})
		if err == nil {
			if _, ok := token.Claims.(*userClaims); ok && token.Valid {
				return http.Cookie{}, false, true
			}
		}
	}
	type refreshClaims struct {
		ID int `json:"id"`
		jwt.StandardClaims
	}
	refreshCookie, err := c.Cookie(currentConfig.Cookie.Refresh.Name)
	if err != nil {
		return http.Cookie{}, false, false
	}
	token, err := jwt.ParseWithClaims(refreshCookie.Value, &refreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return hmacSampleSecret, nil
	})
	if err != nil {
		return http.Cookie{}, false, false
	}
	refresh, ok := token.Claims.(*refreshClaims)
	if !ok || !token.Valid {
		return http.Cookie{}, false, false
	}
	var user models.User
	err = config.DB.Where("id = ?", refresh.ID).First(&user).Error
	if err != nil {
		return http.Cookie{}, false, false
	}
	if token.Raw != user.RefreshToken {
		return http.Cookie{}, false, false
	}
	userToken, _ := createToken(jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(time.Duration(currentConfig.Cookie.User.Expires) * time.Hour).Unix(),
	})
	userNewCookie := http.Cookie{
		Name:   currentConfig.Cookie.User.Name,
		Value:  userToken,
		Path:   "/",
		MaxAge: int((time.Duration(currentConfig.Cookie.User.Expires) * time.Hour).Seconds()),
	}
	return userNewCookie, true, true
}

func getCurrentUser(c echo.Context) (models.User, error) {
	cookie, err := c.Cookie(currentConfig.Cookie.User.Name)
	if err != nil {
		fmt.Println("No cookie")
		return models.User{}, fmt.Errorf("Couldn't find cookie")
	}
	type customClaims struct {
		Email string `json:"email"`
		ID    int    `json:"id"`
		jwt.StandardClaims
	}
	token, err := jwt.ParseWithClaims(cookie.Value, &customClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return hmacSampleSecret, nil
	})
	if err != nil {
		return models.User{}, err
	}
	if claims, ok := token.Claims.(*customClaims); ok && token.Valid {
		var user models.User
		err = nil
		err = config.DB.Where("email = ?", claims.Email).First(&user).Error
		return user, err
	} else {
		return models.User{}, fmt.Errorf("Invalid token")
	}
}

func registerUser(name string, email string, desc string, gender string, dept string) error {
	if len(name) == 0 || len(email) == 0 || len(desc) == 0 || len(gender) == 0 || len(dept) == 0 {
		return fmt.Errorf("Invalid form values")
	}
	var genderEnum models.Gender
	if gender == "male" {
		genderEnum = models.Male
	} else {
		genderEnum = models.Female
	}
	user := models.CreateNewUser(email, name, genderEnum)
	user.Department = dept
	user.Description = desc
	err := config.DB.Save(&user).Error
	return err
}
