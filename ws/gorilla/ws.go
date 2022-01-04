package gorilla

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/delta/orientation-backend/config"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// upgrader configuration to upgrade
// http request to websocket connection
var upgrader = websocket.Upgrader{
	ReadBufferSize:  10240,
	WriteBufferSize: 10240,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

/*
	TODO
	- use here option similar to wap, when user try to connect in multiple tabs
*/
func WsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("odqhowdhqhndiuhnusadyutasbdtyuaSTFYUVASYUDVfuyvasfsyuvtasuyfvtuyastvfduya")
	l := config.Log.WithFields(logrus.Fields{"method": "gorilla/wsHandler"})

	params := r.URL.Query().Get("id")

	userId, err := strconv.Atoi(params)

	if err != nil {
		// handle error
		fmt.Println(err)
		return
	}

	l.Infof("client %s requested websocket connection", userId)

	// l.Debugf("getting user data from cookie")
	// getting authenticated user from the request
	// user, err := auth.GetCurrentUser_http(r)

	// if err != nil {
	// 	l.Errorf("unable to retrive user data from request %+v", err)
	// 	return
	// }

	// l.Infof("Upgarding %s user request to ws connection", user.Username)
	// upgradring http request to websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)

	// closing the request, if there is any error in the upgradation
	if err != nil {
		l.Errorf("error upgrading to webscoket connection %+v", err)
		return
	}

	client := &client{
		id:     userId,
		name:   params,
		wsConn: conn,
	}

	// check if user already established connection
	userRooms.RLock()

	_, ok := userRooms.userRoom[userId]

	if ok {
		response := &responseMessage{
			MessageType: "already-connected",
			Data:        "user already an established connection with the server",
		}
		respJson, _ := json.Marshal(response)
		conn.WriteMessage(websocket.TextMessage, respJson)

		l.Errorf("%s user already have an established connection with the server", params)

		// closing the ws connection
		return
	}

	userRooms.RUnlock()

	// // save username in redis for global chat
	// saveUserNameRedis(userId, params)

	// // broadcasting user status (joined here) for global chat
	// broadcastUserConnectionStatus(client.id, true)

	// unary(request -> response) handles all the ws messages
	go unaryController(conn, client, l)

}
