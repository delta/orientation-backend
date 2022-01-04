package main

import (
	"encoding/json"
	"flag"

	"fmt"
	"log"
	"net/url"

	// "os"
	// "os/signal"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
)

// var addr = flag.String("addr", "localhost:8080", "http service address")

var (
	ip                = flag.String("ip", "127.0.0.1:8080", "server IP")
	connections       = flag.Int("conn", 2, "number of websocket connections")
	speed             = flag.Int("speed", 2, "number of requests per second")
	path              = flag.String("path", "/ws", "ws path")
	room              = "Admin"
	posX        int64 = 50
	posY        int64 = 50
	direction         = "back"
)

type userPosition struct {
	X         int64
	Y         int64
	Direction string
}

type registerUserRequestPayload struct {
	Room     string
	Position userPosition
}

// type
type requestMessage struct {
	MessageType string
	Data        interface{}
}

type moveRequest struct {
	Room     string
	Position userPosition
}

func randomUserPosition() userPosition {
	return userPosition{
		X:         rand.Int63n(20) + posX,
		Y:         rand.Int63n(20) + posY,
		Direction: "back",
	}
}

func main() {

	flag.Parse()
	log.SetFlags(0)

	// interrupt := make(chan os.Signal, 1)
	// signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *ip, Path: *path}
	log.Printf("Connecting to : %s", u.String())

	var conns []*websocket.Conn

	registerPayload, _ := json.Marshal(&requestMessage{MessageType: "user-register", Data: registerUserRequestPayload{
		Room: room,
		Position: userPosition{
			X:         posX,
			Y:         posY,
			Direction: direction,
		},
	}})

	for i := 0; i < *connections; i++ {
		// u.Query().Add("id", fmt.Sprintf("%d", i))
		// u.Query().Set("id", fmt.Sprintf("%d", i))

		q := u.Query()
		q.Set("id", fmt.Sprintf("%d", i))
		u.RawQuery = q.Encode()

		log.Printf("Connecting to :: %s", u.String())

		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

		if err != nil {
			log.Println("Failed to connect", i, err)
			break
		}

		conns = append(conns, c)

		defer func() {

			// closing ws connection
			c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
			time.Sleep(time.Second)
			c.Close()
		}()

		c.WriteMessage(websocket.TextMessage, registerPayload)

		log.Printf("Finished initializing %d connections", len(conns))
	}

	// tts := time.Second
	// if *connections > 100 {
	// 	tts = time.Millisecond * 5
	// }
	tts := time.Second / time.Duration(*speed)

	for {
		for i := 0; i < len(conns); i++ {
			time.Sleep(tts)
			conn := conns[i]
			log.Printf("Conn %d sending message", i)
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second*5)); err != nil {
				log.Printf("Failed to receive pong: %v", err)
			}
			position := randomUserPosition()
			payload, _ := json.Marshal(&requestMessage{
				MessageType: "user-move",
				Data:        moveRequest{Room: room, Position: position},
			})
			conn.WriteMessage(websocket.TextMessage, payload)
		}
	}
}
