package ws

import (
	"fmt"
	"log"
	"net/http"

	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/ws/gorilla"
	"github.com/sirupsen/logrus"
)

func WsMain() {

	l := config.Log.WithFields(logrus.Fields{"method": "ws/wsMain"})

	port := config.Config("WS_PORT")
	addr := fmt.Sprintf(":%s", port)

	l.Infof("Starting websocket server in port %s", port)

	gorilla.InitRooms()
	// gobwas.InitRooms()

	go gorilla.RoomBroadcast()
	// go gobwas.RoomBroadcast()

	// gorilla ws implementation
	http.HandleFunc("/gorilla/ws", gorilla.WsHandler)
	// http.HandleFunc("/gobwas/ws", gobwas.WsHandler)

	log.Fatal(http.ListenAndServe(addr, nil))
}
