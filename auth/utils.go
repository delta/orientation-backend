package auth

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

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
