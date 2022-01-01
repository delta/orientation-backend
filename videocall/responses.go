package videocall

type getAccessTokenResponse struct {
	RoomName string `json:"roomName"`
	Token    string `json:"token"`
}
type roomError struct {
	Message string `json:"message"`
}
