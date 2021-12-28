package ws

import (
	"encoding/json"
	"net/http"

	"github.com/delta/orientation-backend/auth"
	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/core"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// upgrader configuration to upgrade
// http request to websocket connection
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func RegisterRoutes(v *echo.Group) {
	v.GET("/ws", wsHandler)
}

/*
	TODO
	- use here option similiar to wap, when user try to connect in multiple tabs
*/
func wsHandler(c echo.Context) error {
	l := config.Log.WithFields(logrus.Fields{"method": "ws/wsHandler"})

	l.Infof("client requested websocket connection")

	l.Debugf("getting user data from cookie")
	// getting authenticated user from the request
	user, err := auth.GetCurrentUser(c)

	if err != nil {
		l.Errorf("unable to retrive user data from request %+v", err)
		return c.JSON(http.StatusUnauthorized, core.ErrorResponse{Message: "User not authenticated"})
	}

	l.Infof("Upgarding %s user request to ws connection", user.Username)
	// upgradring http request to websocket connection
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)

	// closing the request, if there is any error in the upgradation
	if err != nil {
		l.Errorf("error upgrading to webscoket connection %+v", err)
		return c.JSON(http.StatusInternalServerError, core.ErrorResponse{Message: "Error upgrading to websocket connection"})
	}

	client := &client{
		id:     user.ID,
		name:   user.Name,
		wsConn: conn,
	}

	defer func() {
		closeWs(conn, client)
	}()

	// check if user already established connection
	UserRooms.RLock()

	_, ok := UserRooms.UserRoom[user.ID]

	if ok {
		response := &responseMessage{
			MessageType: "already-conncted",
			Data:        "user already an established conncetion with the server",
		}
		respJson, _ := json.Marshal(response)
		conn.WriteMessage(websocket.TextMessage, respJson)

		l.Errorf("%s user already have an established connection with the server", user.Username)

		// closing the ws connection
		return nil
	}

	UserRooms.RUnlock()

	// unary(request -> response) handles all the ws messages
	unaryController(conn, client, l, c)

	return nil
}
