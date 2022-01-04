package gobwas

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"

	goaway "github.com/TwiN/go-away"
	"github.com/delta/orientation-backend/config"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sirupsen/logrus"
)

var (
	errNotFound     = errors.New("key not found in redis")
	errRoomNotFound = errors.New("room not found")
)

// client type, represents the connected user
type client struct {
	id     int
	name   string
	wsConn *net.Conn
}

// user type, represents the saved user in redis
type user struct {
	Id       int
	Position userPosition
}

// register handler, adds the client to the
// connection pool and redis
func (c *client) register(u *registerUserRequest) error {
	l := config.Log.WithFields(logrus.Fields{"method": "gobwas/register"})

	l.Debugf("registering new user %d in %s room", c.id, u.Room)

	if !isRoomExist(u.Room) {
		return errRoomNotFound
	}

	user := newUser(c.id, u.Room, u.Position)

	room := rooms[u.Room]

	// adding user data to redis
	if err := user.upsertUser(c.id); err != nil {
		return err
	}

	// adding user room in userRoom map
	userRooms.Lock()
	userRooms.userRoom[c.id] = u.Room

	// locking connection pool
	room.Lock()

	// add ws connection handler to the pool
	room.pool[c.id] = c.wsConn

	l.Infof("register new user %d in %s room successful", c.id, u.Room)

	// broadcasts new user data to all the connected clients in that room
	broadcastNewuser(user)

	room.Unlock()

	userRooms.Unlock()

	return nil
}

// de-register handler, removes the client from
// the connection pool and redis
func (c *client) deRegister() error {
	l := config.Log.WithFields(logrus.Fields{"method": "gobwas/deRegister"})

	l.Debugf("de-registering user %d from connection pool", c.id)

	user, err := getUser(c.id)

	if err != nil {
		return err
	}

	userRooms.Lock()
	defer userRooms.Unlock()

	userRoom, ok := userRooms.userRoom[c.id]

	if !ok {
		l.Error("error getting user room from userMap")
	}

	room := rooms[userRoom]

	delete(userRooms.userRoom, c.id)
	// deleting client from connection pool
	room.Lock()
	delete(room.pool, c.id)
	room.Unlock()

	// deleting user from redis
	if user.deleteUser(c.id) != nil {
		return err
	}

	go broadcastUserleftRoom(c.id, userRoom)

	l.Infof("de-registering user %d from connection pool successful", c.id)

	return nil
}

// change room handler, changes user room and updates connection pool
func (c *client) changeRoom(cr *changeRoomRequest) error {
	l := config.Log.WithFields(logrus.Fields{"method": "gobwas/changeRoom"})

	l.Debugf("changing user from %s room to %s room", cr.From, cr.To)

	if !(isRoomExist(cr.From) && isRoomExist(cr.To)) {
		return errRoomNotFound
	}
	// get user
	user, err := getUser(c.id)

	if err != nil {
		return err
	}

	userRooms.Lock()

	userOldRoom, ok := userRooms.userRoom[c.id]

	if !ok {
		// this can happen if user try to move before registering
		// or after deregistering
		return fmt.Errorf("user not found in userMap")
	}

	if userOldRoom != cr.From {
		return fmt.Errorf("user %d not exist in %s room", c.id, cr.From)
	}

	fromRoom := rooms[cr.From]
	toRoom := rooms[cr.To]

	// removing connection handler from old room pool
	fromRoom.Lock()
	conn := fromRoom.pool[c.id]
	delete(fromRoom.pool, c.id)
	fromRoom.Unlock()

	// adding client ws connection handler to new room pool
	toRoom.Lock()

	userRooms.userRoom[c.id] = cr.To
	toRoom.pool[c.id] = conn

	// updating user data(position + room)
	user.Position = cr.Position

	// update user data in redis
	user.upsertUser(c.id)

	// broadcasts new user data to all the connected clients in that room
	broadcastNewuser(user)

	toRoom.Unlock()

	userRooms.Unlock()

	l.Infof("changing user from %s room to %s room successful", cr.From, cr.To)

	// broadcast user left broadcast
	go broadcastUserleftRoom(c.id, cr.From)

	return nil
}

// move handler, updates user data(position and direction) in redis
// BUG ALERT: concurrent writes on same websocket handler will panic
// lock the room of the user while send back the mv response
func (c *client) move(m *moveRequest) error {
	l := config.Log.WithFields(logrus.Fields{"method": "gobwas/move"})

	l.Debugf("updating %s user position in room", c.id)
	// checking if user exists in redis storage
	user, err := getUser(c.id)

	if err != nil {
		return err
	}

	user.Position = m.Position

	userRooms.RLock()
	defer userRooms.RUnlock()

	userRoom := userRooms.userRoom[user.Id]

	room := rooms[userRoom]

	room.Lock()
	defer room.Unlock()

	mvResponse := moveResponse{
		Status: 1,
	}

	response := responseMessage{
		MessageType: "move-response",
		Data:        mvResponse,
	}

	// redis is single threaded, its thread safe :)
	if err := user.upsertUser(c.id); err != nil {
		mvResponse.Status = 0
		response.Data = mvResponse

		resposneJson, _ := json.Marshal(response)

		wsutil.WriteServerMessage(*c.wsConn, ws.OpText, resposneJson)

		return err
	}

	resposneJson, _ := json.Marshal(response)

	wsutil.WriteServerMessage(*c.wsConn, ws.OpText, resposneJson)

	l.Infof("updating %s user position in room is successful", c.id)

	return nil
}

func (c *client) message(m string) {
	l := config.Log.WithFields(logrus.Fields{"method": "gorilla/message"})

	l.Debugf("chat message recieved from user %d", c.id)

	// censoring the messages
	message := goaway.Censor(m)

	chatMessage := &chatMessage{
		Message: message,
		User: chatUser{
			UserId: c.id,
			Name:   c.name,
		},
	}

	response := responseMessage{
		MessageType: "chat-message",
		Data:        chatMessage,
	}
	// global broadcast response message
	go globalBroadCast(response, l)

}

// create new user object
func newUser(id int, room string, userPosition userPosition) *user {
	return &user{
		id,
		userPosition,
	}
}

// create or update user in redis
func (u *user) upsertUser(id int) error {
	key := fmt.Sprintf("user:%d", id)
	userString := u.toJSON()

	if err := config.RDB.Set(key, userString, 0).Err(); err != nil {
		return err
	}
	return nil
}

// deletes user form redis
func (u *user) deleteUser(id int) error {
	key := fmt.Sprintf("user:%d", id)
	return config.RDB.Del(key).Err()
}

// convert user object to json string
func (u *user) toJSON() string {
	jsonData, _ := json.Marshal(u)
	return string(jsonData)
}
