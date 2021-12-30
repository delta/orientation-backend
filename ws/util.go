package ws

import (
	"encoding/json"
	"fmt"

	"github.com/delta/orientation-backend/config"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
)

// utility function to close websocket connection
func closeWs(conn *websocket.Conn, c *client) {
	config.Log.Infof("clinet %d connection closed", c.id)
	// deleting client form connection pool
	if err := c.deRegister(); err != nil {
		config.Log.Errorf("error removing user %s from redis", c.id)
	}
	// closing ws
	conn.Close()
}

// retrives user from redis
func getUser(id int) (*user, error) {
	key := fmt.Sprintf("user:%d", id)

	userJSON, err := config.RDB.Get(key).Result()

	if err == redis.Nil {
		return &user{}, errNotFound
	} else if err != nil {
		return &user{}, err
	}

	var user *user
	userByteArray := []byte(userJSON)

	json.Unmarshal(userByteArray, &user)
	return user, nil
}

// utility func to check if room exist in connction pool
func isRoomExist(room string) bool {
	_, exist := rooms[room]
	return exist
}
