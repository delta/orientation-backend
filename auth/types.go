package auth

import (
	"errors"
	"github.com/golang-jwt/jwt"
)

type Cookie_info interface {
	Get_id() int
	Get_email() (string, error)
}

type userClaims struct {
	Email string `json:"email"`
	ID    int    `json:"id"`
	jwt.StandardClaims
}

type refreshClaims struct {
	ID int `json:"id"`
	jwt.StandardClaims
}

func (m userClaims) Get_id() int {
	return m.ID
}

func (m userClaims) Get_email() (string, error) {
	return m.Email, nil
}

func (m refreshClaims) Get_id() int {
	return m.ID
}

func (m refreshClaims) Get_email() (string, error) {
	return "", errors.New("Refresh token has no email")
}
