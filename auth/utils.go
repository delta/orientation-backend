package auth

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/models"
	"github.com/golang-jwt/jwt"
)

func makeRequest(method string, url string, headers map[string]string, encodedData string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, strings.NewReader(encodedData))
	var returnEmptyByte []byte
	if err != nil {
		return returnEmptyByte, err
	}
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	response, err := client.Do(req)
	if err != nil {
		return returnEmptyByte, err
	}
	defer response.Body.Close()
	temp, _ := ioutil.ReadAll(response.Body)
	return temp, nil
}

func encodeQuery(p map[string]string) (url.Values, error) {
	params := url.Values{}
	for key, value := range p {
		params.Add(key, value)
	}
	return params, nil
}

func createToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(hmacSampleSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func checkAuth(r *http.Request) (string, bool) {
	userCookie, err := r.Cookie(currentConfig.Cookie.User.Name)
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
				return "", true
			}
		}
	}
	type refreshClaims struct {
		ID int `json:"id"`
		jwt.StandardClaims
	}
	refreshCookie, err := r.Cookie(currentConfig.Cookie.Refresh.Name)
	if err != nil {
		return "", false
	}
	token, err := jwt.ParseWithClaims(refreshCookie.Value, &refreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return hmacSampleSecret, nil
	})
	if err != nil {
		return "", false
	}
	refresh, ok := token.Claims.(*refreshClaims)
	if !ok || !token.Valid {
		return "", false
	}
	var user models.User
	err = config.DB.Where("id = ?", refresh.ID).First(&user).Error
	if err != nil {
		return "", false
	}
	if token.Raw != user.RefreshToken {
		return "", false
	}
	userToken, _ := createToken(jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(time.Duration(currentConfig.Cookie.User.Expires) * time.Hour).Unix(),
	})
	userNewCookie := fmt.Sprintf("%s=%s; HttpOnly; Max-Age=%d; Path=/",
		currentConfig.Cookie.User.Name,
		userToken,
		int64((time.Duration(currentConfig.Cookie.User.Expires) * time.Hour).Seconds()),
	)
	return userNewCookie, true
}

func getCurrentUser(r *http.Request) (models.User, error) {

	cookie, err := r.Cookie(currentConfig.Cookie.User.Name)
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
