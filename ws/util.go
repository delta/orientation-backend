package ws

import (
	"encoding/json"
	"fmt"

	"github.com/delta/orientation-backend/config"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// utility function to close websocket connection
func closeWs(conn *websocket.Conn, c *client) {
	config.Log.Infof("clinet %d connection closed", c.id)

	go deleteUserNameRedis(c.id)
	// broadcasting user disconnected status for chat
	go broadcastUserConnectionStatus(c.id, false)
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

// utility func to get all the connected (socket-connected) user
func broadCastAllConnectedUsers(conn *websocket.Conn) {

	// user ids slice
	users := make([]chatUser, 0)

	for _, roomPool := range rooms {
		roomPool.RLock()
	}

	for _, roomPool := range rooms {
		for userId := range roomPool.pool {
			var newUser chatUser

			userName, err := getUserNameRedis(userId)

			if err != nil {
				userName = "Anonymous"
			}

			newUser.UserName = userName
			newUser.UserId = userId

			users = append(users, newUser)
		}
	}

	var response responseMessage

	response.MessageType = "users"
	response.Data = users

	responseJson, _ := json.Marshal(response)

	for _, roomPool := range rooms {
		for _, v := range roomPool.pool {
			v.WriteJSON(responseJson)
		}
	}

	for _, roomPool := range rooms {
		roomPool.RUnlock()
	}
}

// utility func to check if room exist in connction pool
func isRoomExist(room string) bool {
	_, exist := rooms[room]
	return exist
}

// utility func to check if user exist in the room connection pool
func isUserExistRoom(room string, id int) bool {
	if isRoomExist(room) {
		UserRooms.RLock()
		defer UserRooms.RUnlock()

		userRoom := UserRooms.UserRoom[id]

		return userRoom == room
	}

	return false
}

// utility func to save user name in redis
func saveUserNameRedis(userId int, userName string) error {
	l := config.Log.WithFields(logrus.Fields{"method": "ws/util/saveUserNameRedis"})
	key := fmt.Sprintf("username:%d", userId)

	if err := config.RDB.Set(key, userName, 0).Err(); err != nil {
		l.Errorf("error saving %d user's name in redis, %+v", userId, err)
		return err
	}

	return nil
}

// utility func to get username in redis
func getUserNameRedis(userId int) (string, error) {
	key := fmt.Sprintf("username:%d", userId)

	userName, err := config.RDB.Get(key).Result()

	if err == redis.Nil {
		return "", errNotFound
	} else if err != nil {
		return "", err
	}

	return userName, nil
}

// utility func to delete username in redis
func deleteUserNameRedis(userId int) error {
	key := fmt.Sprintf("username:%d", userId)
	return config.RDB.Del(key).Err()
}

func broadcastUserConnectionStatus(userId int, status bool) {

	userName, err := getUserNameRedis(userId)

	if err != nil {
		userName = "Anonymous"
	}

	chatUser := chatUser{
		UserId:   userId,
		UserName: userName,
	}

	response := responseMessage{
		MessageType: "user-connection-status",
		Data: userConnectionStatus{
			Status: status,
			User:   chatUser},
	}

	go globalBroadCast(response)
}
