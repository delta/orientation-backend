package videocall

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	appAuth "github.com/delta/orientation-backend/auth"
	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/core"
	// "github.com/delta/orientation-backend/ws"
	"github.com/labstack/echo/v4"
	"github.com/livekit/protocol/auth"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func randString(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func RegisterRoutes(v *echo.Group) {
	v.GET("/joinvc", JoinVc)
}

func GetJoinToken(apiKey, apiSecret, room, identity string) (string, error) {
	at := auth.NewAccessToken(apiKey, apiSecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     room,
	}
	at.AddGrant(grant).
		SetIdentity(identity).
		SetValidFor(time.Hour)

	return at.ToJWT()
}

func JoinVc(c echo.Context) error {
	apiKey := config.Config("LIVEKIT_KEY")
	apiSecret := config.Config("LIVEKIT_SECRET")
	user, err := appAuth.GetCurrentUser(c)

	if err != nil {
		return c.JSON(http.StatusBadRequest, core.ErrorResponse{Message: "User not authenticated"})
	}
	roomName := c.QueryParam("room")
	fmt.Println("RoomName", roomName)
	randRoomName := randString(40)
	if roomName == "" {
		config.RDB.Set(randRoomName, true, 0)
		roomName = randRoomName
	} else {
		_, err := config.RDB.Get(roomName).Result()
		if err != nil {
			return c.JSON(http.StatusBadRequest, roomError{Message: "Room doesn't exist"})
		} 
	}
	token, err := GetJoinToken(apiKey, apiSecret, roomName, strconv.Itoa(user.ID))
	return c.JSON(http.StatusOK, getAccessTokenResponse{RoomName: roomName, Token: token})
}
