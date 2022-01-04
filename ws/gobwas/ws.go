package gobwas

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/delta/orientation-backend/config"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sirupsen/logrus"
)

/*
	TODO
	- use here option similar to wap, when user try to connect in multiple tabs
*/
func WsHandler(w http.ResponseWriter, r *http.Request) {
	l := config.Log.WithFields(logrus.Fields{"method": "gobwas/wsHandler"})

	params := r.URL.Query().Get("id")

	userId, err := strconv.Atoi(params)

	if err != nil {
		// handle error
		fmt.Println(err)
		return
	}

	// l.Infof("client requested websocket connection")

	// l.Debugf("getting user data from cookie")
	// // getting authenticated user from the request
	// user, err := auth.GetCurrentUser_http(r)

	// if err != nil {
	// 	l.Errorf("unable to retrive user data from request %+v", err)
	// 	return
	// }

	l.Infof("Upgarding %s user request to ws connection", params)
	// upgrading http request to websocket connection
	conn, _, _, err := ws.UpgradeHTTP(r, w)

	// closing the request, if there is any error in the upgradation
	if err != nil {
		l.Errorf("error upgrading to websocket connection %+v", err)
		return
	}

	client := &client{
		id:     userId,
		name:   params,
		wsConn: &conn,
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
		wsutil.WriteServerMessage(conn, ws.OpText, respJson)

		l.Errorf("%s user already have an established connection with the server", params)

		// closing the ws connection
		return
	}

	userRooms.RUnlock()

	// save username in redis for global chat
	saveUserNameRedis(userId, params)

	// broadcasting user status (joined here) for global chat
	broadcastUserConnectionStatus(client.id, true)

	// unary(request -> response) handles all the ws messages
	go unaryController(conn, client, l)
}
