package gorilla

// this file will contain all the request and response type
// of the message passed from and to in the ws connection

// request message type,
// data send from the client to server
// in socket communication
// request message-type -> `regsiter-user`,`change-room`, `user-move`, `chat-message`
type requestMessage struct {
	MessageType string
	Data        map[string]interface{}
}

// response message type
// data send from the server to clinet
// in socket communication
// response message-type -> `room-broadcast`, `new-user`, `already-connected`, `user-left`, `chat-message`, `users`, `user-action`
type responseMessage struct {
	MessageType string
	Data        interface{}
}

// user position(coordinates + direction) type
type userPosition struct {
	X         int64
	Y         int64
	Direction string
}

// register user request type
type registerUserRequest struct {
	Room     string
	Position userPosition
}

// change room(map) request type
type changeRoomRequest struct {
	From     string
	To       string
	Position userPosition
}

// user move position request type
type moveRequest struct {
	Room     string
	Position userPosition
}

// user move response
// status - 1 -> success
// status - 0 -> failed
// client will send move request one - by -one after successful update
type moveResponse struct {
	Status int `json:"status"`
}

// global chat message struct
type chatMessage struct {
	Message string   `json:"message"`
	User    chatUser `json:"user"`
}

// user connection status for global chat
type userConnectionStatus struct {
	Status bool     `json:"status"`
	User   chatUser `json:"user"`
}

// user type in chat
type chatUser struct {
	UserId int    `json:"id"`
	Name   string `json:"name"`
}

type chatRequestMessage struct {
	Message string
}
