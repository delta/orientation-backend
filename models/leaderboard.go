package models

type LeaderBoard struct {
	gameId int `gorm:"column:miniGameId;not null"`
	userId int `gorm:"column:userid;not null"`
	score  int `gorm:"column:score;not null"`
}

func (LeaderBoard) TableName() string {
	return "LeaderBoard"
}
