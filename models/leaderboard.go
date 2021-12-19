package models

type LeaderBoard struct {
	GameId int `gorm:"column:miniGameId;primary_key"`
	UserId int `gorm:"column:userid;primary_key"`
	Score  int `gorm:"column:score;not null"`
}

func (LeaderBoard) TableName() string {
	return "LeaderBoard"
}
