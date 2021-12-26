package videocall

import (
	"fmt"
	"net/http"
	"time"

	appAuth "github.com/delta/orientation-backend/auth"
	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/ws"
	"github.com/labstack/echo/v4"
	"github.com/livekit/protocol/auth"
)

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
	roomName := ws.GetUserRoom(user.ID)
	token, err := GetJoinToken(apiKey, apiSecret, roomName, user.Username)
	if err != nil {
		fmt.Println(err)
	}
	return c.JSON(http.StatusOK, getAccessTokenResponse{Token: token})
}
