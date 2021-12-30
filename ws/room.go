package ws

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/models"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// this type represents the room and its connection pool
type room struct {
	name string
	pool map[int]*websocket.Conn
	sync.Mutex
}

type userRoomMap struct {
	sync.RWMutex
	UserRoom map[int]string
}

// hashmap contains all the rooms and it's connection pool
var rooms map[string]*room = make(map[string]*room)

var UserRooms *userRoomMap = &userRoomMap{
	UserRoom: make(map[int]string),
}

func InitRooms() {
	// fetching list of rooms from db
	roomList, err := models.GetAllRooms()

	if err != nil {
		panic(fmt.Errorf("error fetching rooms from db %+v", err))
	}

	if len(roomList) == 0 {
		config.Log.Infoln("no rooms to connect and broadcast")
	}
	// initializing room connection pool map
	for _, value := range roomList {
		room := &room{
			name: value,
			pool: make(map[int]*websocket.Conn),
		}

		rooms[value] = room
	}
}

// broadcasts users postion to all the rooms(respectively) every **x** seconds
func RoomBroadcast() {
	l := config.Log.WithFields(logrus.Fields{"method": "ws/RoomBroadcast"})

	l.Debug("Starting room broadcasts")

	// x => tickRate, broadcasting frequency (no of requests per second)
	x, _ := strconv.ParseFloat(config.Config("TICK_RATE"), 64)
	var seconds float64 = 1e3 / x

	for _, v := range rooms {
		go func(r *room) {
			for {
				r.roomBroadcast()
				time.Sleep(time.Duration(seconds * 1e6))
			}
		}(v)
	}
}

// get all users of that romm from redis, this method is not **thread safe**
func (r *room) getRoomUsers() ([]string, error) {
	l := config.Log.WithFields(logrus.Fields{"method": "ws/getRoomUsers"})

	l.Debugf("Fetching all users of %s room", r.name)

	keys := make([]string, 0, len(r.pool))
	var users []string

	for k := range r.pool {
		keys = append(keys, fmt.Sprintf("user:%d", k))
	}

	if len(keys) == 0 {
		return users, nil
	}

	// fetching all the users of the room
	u, err := config.RDB.MGet(keys...).Result()

	if err == redis.Nil {
		return users, nil
	} else if err != nil {
		return users, err
	}

	for _, v := range u {
		if v != nil {
			users = append(users, v.(string))
		}
	}

	return users, nil
}

// broadcast users postions in a room to all the clients
// in the connection pool
func (r *room) roomBroadcast() {
	l := config.Log.WithFields(logrus.Fields{"method": "ws/roomBroadcast"})

	l.Debugf("Broadcasting users data to %s room", r.name)

	r.Lock()
	defer r.Unlock()
	users, err := r.getRoomUsers()

	if err != nil {
		l.Errorf("error fetching all users from %s room %+v", r.name, err)
		return
	}

	if len(users) == 0 {
		l.Infof("no users in connection pool - %s room", r.name)
		return

	}

	broadcastData := &responseMessage{
		MessageType: "room-broadcast",
		Data:        users,
	}

	broadcastJsonData, _ := json.Marshal(broadcastData)

	for _, v := range r.pool {
		v.WriteMessage(websocket.TextMessage, broadcastJsonData)
	}

	l.Infof("Broadcast successful for %s room", r.name)
}

// broadcast the newly joined user data
// to all the clients in the room connection pool
// **not thread safe**
func broadcastNewuser(user *user) {
	l := config.Log.WithFields(logrus.Fields{"method": "ws/broadcastNewuser"})
	userRoom, ok := UserRooms.UserRoom[user.Id]

	if !ok {
		l.Error("error getting user room from userMap")
	}

	room := rooms[userRoom]

	response := responseMessage{
		MessageType: "new-user",
		Data:        *user,
	}

	responseJson, _ := json.Marshal(response)

	for _, v := range room.pool {
		v.WriteMessage(websocket.TextMessage, responseJson)
	}

	l.Infof("broadcast new user to %s room is successful", room.name)
}

func broadcastUserleftRoom(userId int, leftRoom string) {
	l := config.Log.WithFields(logrus.Fields{"method": "ws/broadcastUserleftRoom"})

	room := rooms[leftRoom]

	room.Lock()
	defer room.Unlock()

	response := responseMessage{
		MessageType: "user-left",
		Data:        userId,
	}

	responseJson, _ := json.Marshal(response)

	for _, v := range room.pool {
		v.WriteMessage(websocket.TextMessage, responseJson)
	}

	l.Infof("broadcast user left successful for %s room", leftRoom)

}
