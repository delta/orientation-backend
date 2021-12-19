package auth

import (
	"bytes"
	"errors"

	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang-jwt/jwt"
	logger "github.com/sirupsen/logrus"
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
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)

	// non 2XX respones are not considered as error by golang
	if response.StatusCode != 200 {
		return returnEmptyByte, errors.New("request failed")
	}

	defer response.Body.Close()
	// temp, _ := ioutil.ReadAll(response.Body)
	// fmt.Println("\n", temp)
	return buf.Bytes(), nil
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
	tokenString, err := token.SignedString(HmacSampleSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func get_token(token *jwt.Token) (interface{}, error) {
	l := logger.WithFields(logger.Fields{
		"method": "auth/utils/get_token",
	})
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		errMsg := fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"])
		l.Errorf(errMsg)
		return nil, errors.New(errMsg)
	}

	return HmacSampleSecret, nil
}

func check_claim(token *jwt.Token, ok bool) bool {
	l := logger.WithFields(logger.Fields{
		"method": "auth/utils/check_claim",
	})
	if !ok || !token.Valid {
		l.Errorf("Couldn't verify token")
		return false
	}
	return true
}

func Get_info_from_cookie(cookie *http.Cookie, cookie_name string) (Cookie_info, error) {
	l := logger.WithFields(logger.Fields{
		"method":     "auth/utils/Get_user_from_cookie",
		"userCookie": cookie,
	})
	var token *jwt.Token
	var err error
	l.Infof("Parsing cookie")
	if cookie_name == "user" {
		token, err = jwt.ParseWithClaims(cookie.Value, &userClaims{}, get_token)
	} else if cookie_name == "refresh" {
		token, err = jwt.ParseWithClaims(cookie.Value, &refreshClaims{}, get_token)
	} else {
		l.Errorf("Invalid cookie name")
		return nil, errors.New("Invalid cookie name")
	}
	if err != nil {
		l.Errorf("Error parsing cookie")
		return nil, err
	}
	if cookie_name == "user" {
		data, ok := token.Claims.(*userClaims)
		if !check_claim(token, ok) {
			return nil, errors.New("Cannot verify claims")
		}
		fmt.Println(data)
		return data, nil
	} else if cookie_name == "refresh" {
		data, ok := token.Claims.(*refreshClaims)
		if !check_claim(token, ok) {
			return nil, errors.New("Cannot verify claims")
		}
		return data, nil
	}
	return nil, errors.New("Unknown error")
}
