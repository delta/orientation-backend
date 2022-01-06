package ws

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// websocket unary handler, reads request message
// from the websocket connection i.e the client and
// and responds respectively
func unaryController(conn *websocket.Conn, client *client, l *logrus.Entry) error {

	defer func() {
		closeWs(conn, client)
	}()

	for {
		fmt.Println("socket reading")
		// reads the message
		_, p, err := conn.ReadMessage()

		if err != nil {
			l.Errorf("Error reading from socket connection %+v", err)
			return nil
		}

		var requestMessage requestMessage

		if err := json.Unmarshal(p, &requestMessage); err != nil {
			l.Errorf("error parsing request message %v", err)
			return nil
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
				l.Errorf("error registering user %s in %s room, err : %+v", client.id, registerUserRequest.Room, err)
				return nil
			}
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
				return nil
			}
		/*
			`change-room` updates user room in redis and
			connection pool, broadcasts `new-user` similar to
			user-register
		*/
		case "change-room":
			reqJson, _ := json.Marshal(requestMessage.Data)
			var changeRoomRequest changeRoomRequest
			json.Unmarshal(reqJson, &changeRoomRequest)

			if err := client.changeRoom(&changeRoomRequest); err != nil {
				l.Errorf("error changing user %s room %+v", client.id, err)
				return nil
			}

		case "chat-message":
			reqJson, _ := json.Marshal(requestMessage.Data)
			var chatRequest chatRequest
			json.Unmarshal(reqJson, &chatRequest)

			client.message(&chatRequest)

		default:
			l.Debugln("Invalid socket request message type")
			// closing the ws connection
			return nil
		}
	}
}
