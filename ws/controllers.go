package ws

import (
	"encoding/json"
	"net/http"

	"github.com/delta/orientation-backend/core"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// websocket unary handler, reads request message
// from the websocket connection i.e the client and
// and responds respectively
func unaryController(conn *websocket.Conn, client *client, l *logrus.Entry, c echo.Context) error {
	for {
		// reads the message
		_, p, err := conn.ReadMessage()

		if err != nil {
			l.Errorf("Error readig from socket connection %+v", err)
			return c.JSON(http.StatusInternalServerError, core.ErrorResponse{Message: "error reading from socket connection"})
		}

		var requestMessage requestMessage

		if err := json.Unmarshal(p, &requestMessage); err != nil {
			l.Errorf("error parsing request message %v", err)
			return c.JSON(http.StatusInternalServerError, core.ErrorResponse{Message: "error parsing socket message"})
		}

		switch requestMessage.MessageType {
		/*
			`register-user` type request message
			adds user to connection pool and in redis.
			broadcasts `new-user` response message with
			user position to all the connected clients in
			that room.
		*/
		case "user-register":
			reqJson, _ := json.Marshal(requestMessage.Data)
			var registerUserRequest registerUserRequest
			json.Unmarshal(reqJson, &registerUserRequest)

			if err := client.register(&registerUserRequest); err != nil {
				l.Errorf("error registering user %s in %s room", client.id, registerUserRequest.Room)
				return c.JSON(http.StatusInternalServerError, core.ErrorResponse{Message: "error registering user"})
			}
			user := newUser(client.name, registerUserRequest.Room, registerUserRequest.Position)
			// broadcasts new user data to all the connected clients
			broadcastNewuser(user)
		/*
			`user-move` type request message
			updates user position in redis.
		*/
		case "user-move":
			reqJson, _ := json.Marshal(requestMessage.Data)
			var moverequest moveRequest
			json.Unmarshal(reqJson, &moverequest)

			if err := client.move(&moverequest); err != nil {
				l.Errorf("error updating user %d position in redis : %+v", client.id, err)
				return c.JSON(http.StatusInternalServerError, core.ErrorResponse{Message: "error updating user data"})

			}
		/*
			`change-room` updates user room in redis and
			connection pool, broadcasts `new-user` similair to
			user-register
		*/
		case "change-room":
			reqJson, _ := json.Marshal(requestMessage.Data)
			var changeRoomRequest changeRoomRequest
			json.Unmarshal(reqJson, &changeRoomRequest)

			if err := client.changeRoom(&changeRoomRequest); err != nil {
				l.Errorf("error changing user %s room %+v", client.id, err)
				return c.JSON(http.StatusInternalServerError, core.ErrorResponse{Message: "error updating user room"})
			}
			user := newUser(client.name, changeRoomRequest.To, changeRoomRequest.Position)
			// broadcasts updated user data to all the connected clients
			broadcastNewuser(user)

		default:
			l.Debugln("Invalid socket request message type")
			return c.JSON(http.StatusBadRequest, core.ErrorResponse{Message: "invalid request message type"})
		}
	}
}
