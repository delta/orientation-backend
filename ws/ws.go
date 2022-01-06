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

	l.Infof("Upgarding %s user request to ws connection", user.Name)
	// upgrading http request to websocket connection
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)

	// closing the request, if there is any error in the upgradation
	if err != nil {
		l.Errorf("error upgrading to webscoket connection %+v", err)
		return c.JSON(http.StatusInternalServerError, core.ErrorResponse{Message: "Error upgrading to websocket connection"})
	}

	client := &client{
		id:     user.ID,
		name:   user.Username,
		wsConn: conn,
	}

	// check if user already established connection
	// if yes,remove and close the old connection
	userRoom, err := getUserRoom(client.id)

	if err == nil {
		l.Errorf("%s user already have an established connection with the server", user.Username)

		response := &responseMessage{
			MessageType: "already-connected",
			Data:        "user already an established connection with the server",
		}

		room := rooms[userRoom]
		// removing user connection from the pool
		room.Lock()
		oldConn := room.pool[user.ID]
		delete(room.pool, user.ID)
		room.Unlock()

		respJson, _ := json.Marshal(response)
		oldConn.WriteMessage(websocket.TextMessage, respJson)
		// closing the old connection
		// and they can continue using this connection
		closeConnection(oldConn, user.ID, l)

		l.Infof("Closed the %s user old connection")
	}

	// unary(request -> response) handles all the ws messages
	go unaryController(conn, client, l)

	return nil
}
