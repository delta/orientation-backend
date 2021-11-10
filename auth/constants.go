package auth

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

var uiURL string = "http://localhost:3000"

var currentConfig = getConfig()

var hmacSampleSecret = []byte(currentConfig.Cookie.Jwt_secret)
