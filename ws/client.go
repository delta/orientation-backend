package ws

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/delta/orientation-backend/config"
	"github.com/gorilla/websocket"
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
	wsConn *websocket.Conn
}

// user type, represents the saved user in redis
type user struct {
	Id       int
	Position userPosition
}

// register handler, adds the client to the
// connection pool and redis
func (c *client) register(u *registerUserRequest) error {
	l := config.Log.WithFields(logrus.Fields{"method": "ws/register"})

	l.Debugf("registering new user %d in %s room", c.id, u.Room)

	if !isRoomExist(u.Room) {
		return errRoomNotFound
	}

	user := newUser(c.id, u.Room, u.Position)

	room, err := getUserRoom(c.id)

	if err == nil {
		return fmt.Errorf("user already registered, now in room %s", room)

	} else if err != errNotFound {
		return err
	}

	if err := saveUserRoom(user.Id, u.Room); err != nil {
		return fmt.Errorf("error saving user %d room to redis", user.Id)
	}

	// adding user data to redis
	if err := user.upsertUser(c.id); err != nil {
		return err
	}

	roomPool := rooms[u.Room]
	// locking connection pool
	roomPool.Lock()
	defer roomPool.Unlock()

	// add ws connection handler to the pool
	roomPool.pool[c.id] = c.wsConn

	l.Infof("register new user %d in %s room successful", c.id, u.Room)

	return nil
}

// de-register handler, removes the client from
// the connection pool and redis
func (c *client) deRegister() error {
	l := config.Log.WithFields(logrus.Fields{"method": "ws/deRegister"})

	l.Debugf("de-registering user %d from connection pool", c.id)

	user, err := getUser(c.id)

	if err != nil {
		return err
	}

	room, err := getUserRoom(c.id)

	if err != nil {
		return fmt.Errorf("unable to get user room from redis %+v", err)
	}

	if err := deleteUserRoom(user.Id); err != nil {
		l.Errorf("error deleting user %d room in redis", user.Id)
	}

	roomPool := rooms[room]

	// deleting client from connection pool
	roomPool.Lock()
	defer roomPool.Unlock()
	delete(roomPool.pool, c.id)

	// deleting user from redis
	if user.deleteUser(c.id) != nil {
		return err
	}

	l.Infof("de-registering user %d from connection pool successful", c.id)

	broadcastUserleftRoom(c.id, room)

	return nil
}

// change room handler, changes user room and updates connection pool
func (c *client) changeRoom(cr *changeRoomRequest) error {
	l := config.Log.WithFields(logrus.Fields{"method": "ws/changeRoom"})

	l.Debugf("changing user from %s room to %s room", cr.From, cr.To)

	if !(isRoomExist(cr.From) && isRoomExist(cr.To)) {
		return errRoomNotFound
	}
	// get user
	user, err := getUser(c.id)

	if err != nil {
		return err
	}

	fromRoom := rooms[cr.From]
	toRoom := rooms[cr.To]

	// removing connection handler from old room pool
	fromRoom.Lock()
	conn := fromRoom.pool[c.id]
	delete(fromRoom.pool, c.id)
	fromRoom.Unlock()

	// updating user data(position + room)
	user.Position = cr.Position

	// update user data in redis
	user.upsertUser(c.id)

	// adding client ws connection handler to new room pool
	toRoom.Lock()
	toRoom.pool[c.id] = conn

	// broadcasts new user data to all the connected clients in that room
	// broadcastNewuser(user)

	broadcastUserleftRoom(c.id, cr.From)

	toRoom.Unlock()

	l.Infof("changing user from %s room to %s room successful", cr.From, cr.To)

	return nil
}

// move handler, updates user data(position and direction) in redis
func (c *client) move(m *moveRequest) error {
	l := config.Log.WithFields(logrus.Fields{"method": "ws/move"})

	l.Debugf("updating %s user position in room", c.id)
	// checking if user exists in redis storage
	user, err := getUser(c.id)

	if err != nil {
		return err
	}

	user.Position = m.Position

	// redis is single threaded, its thread safe :)
	if err := user.upsertUser(c.id); err != nil {
		return err
	}

	l.Infof("updating %s user position in room is successful", c.id)

	return nil
}

func (c *client) message(ch *chatRequest) error {
	l := config.Log.WithFields(logrus.Fields{"method": "ws/message"})

	l.Debugf("trying to global broadcast the message form user %s", c.name)

	chatResponse := chatResponse{
		Message:  ch.Message,
		UserName: c.name,
	}

	res := responseMessage{
		MessageType: "chat-message",
		Data:        chatResponse,
	}

	// globally broadcasting message to all the users
	go globalBroadcast(res)

	return nil

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
