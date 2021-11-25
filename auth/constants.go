package auth

// var totalSprites = 4
import "github.com/golang-jwt/jwt"

type CustomClaims struct {
	Email string `json:"email"`
	ID    int    `json:"id"`
	jwt.StandardClaims
}

type ErrorResponse struct {
	Message string `json:"message"`
	Error   error  `json:"error"`
}

var uiURL string = "http://localhost:3000"

// Dauth Config
var CurrentConfig = getConfig()

var HmacSampleSecret = []byte(CurrentConfig.Cookie.Jwt_secret)
