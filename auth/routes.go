package auth

import (
	"encoding/json"
	"fmt"
	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/models"
	"github.com/golang-jwt/jwt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var totalSprites = 4

type TokenResult struct {
	Type    string `json:"token_type"`
	Token   string `json:"access_token"`
	State   string `json:"state"`
	Expires int64  `json:"expires_in"`
}

type UserResult struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	Gender string `json:"gender"`
}

type isAuthResult struct {
	Status bool `json:"status"`
}

var currentConfig = getConfig()

var hmacSampleSecret = []byte(currentConfig.Cookie.Jwt_secret)

func Auth(w http.ResponseWriter, r *http.Request) {
	user, err := getCurrentUser(r)
	fmt.Println(user)
	fmt.Println(err)
	if err == nil {
		http.Redirect(w, r, "http://localhost:3000", 302)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	if err != nil {
		fmt.Println("Inside new auth")
		// fmt.Println(r.Host)
		params := map[string]string{
			"client_id":     currentConfig.Dauth.Client_id,
			"redirect_uri":  config.Config("HOME_PAGE_URI") + "/auth/callback",
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
		http.Redirect(w, r, base.String(), 302)
	} else {
		// fmt.Println("Hel")
		http.Redirect(w, r, "http://localhost:3000/", 302)
	}
}

func CallBack(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query()["code"][0]

	// Getting Token
	params := map[string]string{
		"code":          code,
		"client_secret": currentConfig.Dauth.Client_secret,
		"client_id":     currentConfig.Dauth.Client_id,
		"redirect_uri":  config.Config("HOME_PAGE_URI") + "/auth/callback",
		"grant_type":    currentConfig.Dauth.Grant_type}
	tokenQueryString, err := encodeQuery(params)
	if err != nil {
		panic(fmt.Errorf("Error with url"))
	}
	tokenEncodedData := tokenQueryString.Encode()
	tokenHeader := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	tokenResponse, err := makeRequest(http.MethodPost, "https://auth.delta.nitt.edu/api/oauth/token", tokenHeader, tokenEncodedData)
	if err != nil {
		return
	}
	var tokenResult TokenResult
	err = json.Unmarshal(tokenResponse, &tokenResult)
	if err != nil {
		http.Redirect(w, r, "http://localhost:3000", 302)
		return
	}

	// User resouce
	resourceHeader := map[string]string{"Authorization": "Bearer " + tokenResult.Token}
	userResponse, err := makeRequest(http.MethodPost, "https://auth.delta.nitt.edu/api/resources/user", resourceHeader, url.Values{}.Encode())
	if err != nil {
		http.Redirect(w, r, "http://localhost:3000", 302)
		return
	}
	var userResult UserResult
	err = json.Unmarshal(userResponse, &userResult)
	if err != nil {
		http.Redirect(w, r, "http://localhost:3000", 302)
		return
	}
	// isNewUser := false

	// Creating user
	var user models.User
	var gender models.Gender
	if userResult.Gender == "MALE" {
		gender = models.Male
	} else {
		gender = models.Female
	}
	if err = config.DB.Where("email = ?", userResult.Email).First(&user).Error; err != nil {
		rand.Seed(time.Now().UnixNano())
		user = models.User{Email: userResult.Email, Name: userResult.Name, Gender: gender, SpriteSheetID: rand.Intn(totalSprites) + 1}
		// isNewUser = true
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
	http.SetCookie(w, &userCookie)
	http.SetCookie(w, &refreshCookie)
	http.Redirect(w, r, "http://localhost:3000", 302)
}

func LogOut(w http.ResponseWriter, r *http.Request) {
	userCookie := fmt.Sprintf("%s=; HttpOnly; Max-Age=%d; Path=/",
		currentConfig.Cookie.User.Name,
		-1,
	)
	w.Header().Add("Set-Cookie", userCookie)
	refreshCookie := fmt.Sprintf("%s=; HttpOnly; Max-Age=%d; Path=/",
		currentConfig.Cookie.Refresh.Name,
		-1,
	)
	w.Header().Add("Set-Cookie", refreshCookie)
	json.NewEncoder(w).Encode(isAuthResult{Status: true})
}

func CheckAuth(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered")
	newUserCookie, result := checkAuth(r)
	if newUserCookie != "" {
		w.Header().Add("Set-Cookie", newUserCookie)
	}
	json.NewEncoder(w).Encode(isAuthResult{Status: result})
}
