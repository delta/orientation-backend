package leaderboard

import (
	"fmt"
	"net/http"

	"strconv"

	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/models"
	"github.com/labstack/echo/v4"
	logger "github.com/sirupsen/logrus"
)

func RegisterRoutes(v *echo.Group) {
	v.POST("/addscore", addScore)
	v.GET("/leaderboard:minigameId", getLeaderBoard)
}

type Score struct {
	Game  string `json:"game"`
	Score string `json:"score"`
}

type Leaderboard struct {
	Name       int    `gorm:"column:Name;" json:"userId"`
	Score      int    `gorm:"column:Score;" json:"score`
	Department string `gorm:"column:Department;" json:"department"`
}

func addScore(c echo.Context) error {
	l := logger.WithFields(logger.Fields{
		"method": "leaderboard/routes/AddScore",
	})
	s := new(Score)
	if err := c.Bind(s); err != nil {
		l.Errorf("Incorrect data sent")
		return err
	}
	fmt.Println(s.Game)
	sc, _ := strconv.Atoi(s.Score)
	err := handleAddScore(c, s.Game, sc)
	if err != nil {

		l.Errorf("Couldn't add score")
		return c.JSON(http.StatusBadRequest, ScoreAddStatus{Status: false, Message: "Couldn't add score"})
	}
	return c.JSON(http.StatusOK, ScoreAddStatus{Status: true, Message: "Added successfully"})
}

func getLeaderBoard(c echo.Context) error {
	l := logger.WithFields(logger.Fields{
		"method": "leaderboard/routes/getLeaderBoard",
	})

	l.Infoln("getleaderBoard requested")

	minigameId := c.Param("minigameId")

	var response []Leaderboard

	db := config.DB

	var miniGame models.MiniGame

	if err := db.First(&miniGame, minigameId).Error; err != nil {
		l.Errorf("Error fetching miniGame from db %+v", err)
		return c.JSON(http.StatusBadRequest, LeaderBoardResponse{Leaderboard: response, Message: "invalid minigame"})
	}

	query := fmt.Sprintf("SELECT u.userName AS Name,u.department AS Department,l.score AS Score FROM LeaderBoard AS l LEFT JOIN User AS u ON l.userId = u.id WHERE l.miniGameId = %s ORDER BY l.score DESC;", minigameId)

	if err := db.Raw(query).Scan(&response).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, LeaderBoardResponse{Leaderboard: response, Message: "internal server error"})
	}

	l.Infoln("getleaderBoard for minigame id %s is successful", minigameId)

	return nil
}
