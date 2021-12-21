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
	logger "github.com/sirupsen/logrus"
)

// ands the auth code to the database,
// and creates an user if the user doesn't exist in the database
// returns userToken, refreshToken, user, isNewUser(bool), error
func handleCallBack(code string) (string, string, models.User, bool, error) {
	var l = logger.WithFields(logger.Fields{
		"method": "auth/controllers/handleCallback",
		"code":   code, // is it okay to log the the code ?
	})
	l.Infof("a user trying to login with dauth login code")
	params := map[string]string{
		"code":          code,
		"client_secret": CurrentConfig.Dauth.Client_secret,
		"client_id":     CurrentConfig.Dauth.Client_id,
		"redirect_uri":  CurrentConfig.Dauth.Redirect_uri,
		"grant_type":    CurrentConfig.Dauth.Grant_type,
	}
	tokenQueryString, err := encodeQuery(params)
	if err != nil {
		l.Errorf(("Couldn't encode error, due to %v"), err)
		// TODO: shd do proper error handling instead of using panic
		panic(fmt.Errorf("error with url"))
	}

	l.Debugf("Attempting to get user token from dauth")
	tokenEncodedData := tokenQueryString.Encode()
	tokenHeader := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	tokenResponse, err := makeRequest(http.MethodPost, "https://auth.delta.nitt.edu/api/oauth/token", tokenHeader, tokenEncodedData)
	if err != nil {
		l.Errorf("Fetching user token resulted in an error, %v", err)
		return "", "", models.User{}, false, err
	}
	var tokenResult TokenResult

	// l.Debugf("Got the response %v when trying to fetch user token", tokenResponse)
	err = json.Unmarshal(tokenResponse, &tokenResult)
	if err != nil {
		l.Errorf("Unable to unmarshal the the user token")
		return "", "", models.User{}, false, err
	}
	l.Debugf("Got the response %v when trying to fetch user token", tokenResult)

	l.Debugf("trying to fetch user data using user token with %v", tokenResult.Token)

	resourceHeader := map[string]string{"Authorization": "Bearer " + tokenResult.Token}
	userResponse, err := makeRequest(http.MethodPost, "https://auth.delta.nitt.edu/api/resources/user", resourceHeader, url.Values{}.Encode())
	if err != nil {
		l.Errorf("Fetching user data resulted in an error, %v", err)
		return "", "", models.User{}, false, err
	}

	var userResult UserResult
	err = json.Unmarshal(userResponse, &userResult)
	l.Errorf("Unable to unmarshal the the user token")
	if err != nil {
		l.Errorf("Unable to unmarshal the the user response")
		return "", "", models.User{}, false, err
	}
	l.Debugf("Got the response %v when trying to fetch user user data", userResult)
	var gender models.Gender
	if userResult.Gender == "MALE" {
		gender = models.Male
	} else {
		gender = models.Female
	}

	l.Infof("Successfully fetched user data from Dauth. Now checking if a user record exists locally")

	user, isUnsuccessful := models.GetOnCondition("email", userResult.Email)
	isNewUser := false
	if isUnsuccessful {
		isNewUser = true
		l.Debugf("A user record doesn't exist in database, creating one")
		user = models.CreateNewUser(userResult.Email, userResult.Name, gender)
	}
	userToken, _ := createToken(jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(time.Duration(CurrentConfig.Cookie.User.Expires) * time.Hour).Unix(),
	})

	refreshToken, _ := createToken(jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().Add(time.Duration(CurrentConfig.Cookie.Refresh.Expires) * time.Hour).Unix(),
	})
	user.RefreshToken = refreshToken
	config.DB.Save(&user)
	l.Infof("isNewUser=%v, Created new user and refresh token for the user", isNewUser)
	isNewUser = (user.Description == "")
	return userToken, refreshToken, user, isNewUser, nil
}

// Gets the currently logged in user with the cookie
func GetCurrentUser(c echo.Context) (models.User, error) {
	cookie, err := c.Cookie(CurrentConfig.Cookie.User.Name)
	if err != nil {
		fmt.Println("No cookie")
		return models.User{}, fmt.Errorf("couldn't find cookie")
	}
	token, err := jwt.ParseWithClaims(cookie.Value, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return HmacSampleSecret, nil
	})
	if err != nil {
		return models.User{}, err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		var user models.User
		err = nil
		err = config.DB.Where("email = ?", claims.Email).First(&user).Error
		return user, err
	} else {
		return models.User{}, fmt.Errorf("invalid token")
	}
}

func createDummyUser(roll int) (string, string) {
	email := fmt.Sprintf("%d@nitt.edu", roll)
	user, no_user := models.GetOnCondition("email", email)
	if no_user {
		user = models.CreateNewUser(email, "A", models.Male)
	}
	userToken, _ := createToken(jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(time.Duration(CurrentConfig.Cookie.User.Expires) * time.Hour).Unix(),
	})
	refreshToken, _ := createToken(jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().Add(time.Duration(CurrentConfig.Cookie.Refresh.Expires) * time.Hour).Unix(),
	})
	user.RefreshToken = refreshToken
	config.DB.Save(&user)
	return userToken, refreshToken
}
