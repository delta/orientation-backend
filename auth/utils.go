package auth

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

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
