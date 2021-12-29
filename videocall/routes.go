package videocall

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	appAuth "github.com/delta/orientation-backend/auth"
	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/core"
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

	if err != nil {
		return c.JSON(http.StatusInternalServerError, core.ErrorResponse{Message: "User not authenticated"})
	}
	ws.UserRooms.RLock()
	defer ws.UserRooms.RUnlock()

	roomName := ws.UserRooms.UserRoom[user.ID]

	token, err := GetJoinToken(apiKey, apiSecret, roomName, strconv.Itoa(user.ID))
	if err != nil {
		fmt.Println(err)
	}
	return c.JSON(http.StatusOK, getAccessTokenResponse{Token: token})
}
