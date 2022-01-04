package gorilla

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
	config.Log.Infof("client %d connection closed", c.id)

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
// will be broadcast after user registers
func sendAllConnectedUsers(conn *websocket.Conn, userid int) {
	l := config.Log.WithFields(logrus.Fields{"method": "gorilla/util/sendCastAllConnectedUsers"})

	l.Debugf("trying to broadcast all the users to %d user", userid)
	// user ids slice
	users := make([]chatUser, 0)

	for _, roomPool := range rooms {
		roomPool.Lock()
	}

	l.Info("Locked all the rooms to get connected users")

	for _, roomPool := range rooms {
		for userId := range roomPool.pool {
			var newUser chatUser

			userName, err := getUserNameRedis(userId)

			if err != nil {
				userName = "Anonymous"
			}

			newUser.Name = userName
			newUser.UserId = userId

			users = append(users, newUser)
		}
	}

	var response responseMessage

	response.MessageType = "users"
	response.Data = users

	responseJson, _ := json.Marshal(response)

	if err := conn.WriteMessage(websocket.TextMessage, responseJson); err != nil {
		l.Errorf("error writing message %+v", err)
	}

	l.Debugf("request message sent successfull")

	for _, roomPool := range rooms {
		roomPool.Unlock()
	}

	l.Info("UnLocked all the rooms to get connected users")
}

// utility func to check if room exist in connection pool
func isRoomExist(room string) bool {
	_, exist := rooms[room]
	return exist
}

// utility func to save user name in redis
func saveUserNameRedis(userId int, userName string) error {
	l := config.Log.WithFields(logrus.Fields{"method": "gorilla/util/saveUserNameRedis"})
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

// broadcast user connects and disconnects status to all the other connected
// clients **thread safe**
func broadcastUserConnectionStatus(userId int, status bool) {
	l := config.Log.WithFields(logrus.Fields{"method": "gorilla/broadcastUserConnectionStatus"})

	l.Debugf("trying to broadcast %d user connection status", userId)

	userName, err := getUserNameRedis(userId)

	if err != nil {
		userName = "Anonymous"
	}

	chatUser := chatUser{
		UserId: userId,
		Name:   userName,
	}

	response := responseMessage{
		MessageType: "user-action",
		Data: userConnectionStatus{
			Status: status,
			User:   chatUser},
	}

	go globalBroadCast(response, l)
}
