package ws

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/delta/orientation-backend/config"
	"github.com/gorilla/websocket"
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
	Name     string
	Room     string
	Position userPosition
}

// register handler, adds the client to the
// connection pool and redis
func (c *client) register(u *registerUserRequest) error {
	config.Log.Debugf("registering new user %d in %s room", c.id, u.Room)

	if !isRoomExist(u.Room) {
		return errRoomNotFound
	}

	user := newUser(c.name, u.Room, u.Position)

	room := rooms[u.Room]

	// adding user data to redis
	if err := user.upsertUser(c.id); err != nil {
		return err
	}

	// locking connection pool
	room.Lock()
	defer room.Unlock()

	// add ws connection handler to the pool
	room.pool[c.id] = c.wsConn

	return nil
}

// de-register handler, removes the client from
// the connection pool and redis
func (c *client) deRegister() error {
	config.Log.Debugf("de-registering user %d from connection pool", c.id)

	user, err := getUser(c.id)

	if err != nil {
		return err
	}

	room := rooms[user.Room]

	// deleting client from connection pool
	room.Lock()
	defer room.Unlock()
	delete(room.pool, c.id)

	// deleting user from redis
	if user.deleteUser(c.id) != nil {
		return err
	}

	return nil
}

// change room handler, changes user room and updates connection pool
func (c *client) changeRoom(cr *changeRoomRequest) error {
	config.Log.Debugf("changing user from %s room to %s room", cr.From, cr.To)

	if !(isRoomExist(cr.From) && isRoomExist(cr.To)) {
		return errRoomNotFound
	}
	// get user
	user, err := getUser(c.id)

	if err != nil {
		return err
	}

	if !isUserExistRoom(cr.From, c.id) {
		return fmt.Errorf("user %d not exist in %s room", c.id, cr.From)
	}

	fromRoom := rooms[cr.From]
	toRoom := rooms[cr.To]

	// removing connction handler from old room pool
	fromRoom.Lock()
	conn := fromRoom.pool[c.id]
	delete(fromRoom.pool, c.id)
	fromRoom.Unlock()

	// adding client ws connection handler to new room pool
	toRoom.Lock()
	toRoom.pool[c.id] = conn
	toRoom.Unlock()

	// updating user data(position + room)
	user.Room = cr.To
	user.Position = cr.Position

	// update user data in redis
	user.upsertUser(c.id)

	return nil
}

// move handler, updates user data(position and direction) in redis
func (c *client) move(m *moveRequest) error {
	// checking if user exists in redis storage
	user, err := getUser(c.id)

	if err != nil {
		return err
	}

	if !isUserExistRoom(m.Room, c.id) {
		return fmt.Errorf("user %d not exist in %s room", c.id, m.Room)
	}

	user.Name = c.name
	user.Room = m.Room
	user.Position = m.Position

	// redis is single threaded, its thread safe :)
	if err := user.upsertUser(c.id); err != nil {
		return err
	}
	return nil
}

// create new user object
func newUser(name, room string, userPosition userPosition) *user {
	return &user{
		name,
		room,
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
