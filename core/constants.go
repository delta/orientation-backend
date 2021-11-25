package core

type ErrorResponse struct {
	Message string `json:"message"`
	Error   error  `json:"error"`
}
