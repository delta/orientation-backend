package leaderboard

type ScoreAddStatus struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type leaderBoardResponse struct {
	Leaderboard []leaderboard `json:"leaderboard"`
	Message     string        `json:"message"`
}

type leaderboard struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Score      int    `json:"score"`
}
