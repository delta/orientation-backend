package leaderboard

type ScoreAddStatus struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type leaderBoardResponse struct {
	Leaderboard []leaderboard `json:"leaderboard"`
	Message     string        `json:"message"`
}
