package leaderboard

type ScoreAddStatus struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type LeaderBoardResponse struct {
	Leaderboard []Leaderboard `json:"leaderboard"`
	Message     string        `json:"message"`
}
