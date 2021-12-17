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
)

// this type represents the room and its connection pool
type room struct {
	name string
	pool map[int]*websocket.Conn
	sync.Mutex
}

// hashmap contains all the rooms and it's connection pool
var rooms map[string]*room = make(map[string]*room)

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
	// x, broadcasting frequency
	x, _ := strconv.Atoi(config.Config("x"))

	for _, v := range rooms {
		go func(r *room) {
			for {
				r.broadcastUsers()
				time.Sleep(time.Second * time.Duration(x))
			}
		}(v)
	}
}

// get all users of that romm from redis, this method is not **thread safe**
func (r *room) getAllUsers() ([]string, error) {
	config.Log.Debugf("Fetching all users of %s room", r.name)

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
func (r *room) broadcastUsers() {
	config.Log.Debugf("Broadcasting users data to %s room", r.name)

	r.Lock()
	defer r.Unlock()
	users, err := r.getAllUsers()

	if err != nil {
		config.Log.Errorf("error fetching all users from %s room %+v", r.name, err)
		return
	}

	if len(users) == 0 {
		config.Log.Infof("no users in connection pool - %s room", r.name)
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

	config.Log.Infof("Broadcast sucessfull for %s room", r.name)
}

// broadcast the newly joined user data
// to all the clients in the room connection pool
func broadcastNewuser(user *user) {
	room := rooms[user.Room]

	room.Lock()
	defer room.Unlock()

	response := responseMessage{
		MessageType: "new-user",
		Data:        *user,
	}

	responseJson, _ := json.Marshal(response)

	for _, v := range room.pool {
		v.WriteMessage(websocket.TextMessage, responseJson)
	}

	config.Log.Infof("broadcast new user to %s room is sucessful", room.name)
}
