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
	v.GET("/leaderboard/:minigameId", getLeaderBoard)
}

type Score struct {
	Game  string `json:"game"`
	Score string `json:"score"`
}

type leaderboard struct {
	Name       string `json:"name"`
	Score      int    `json:"score"`
	Department string `json:"department`
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

	minigameChar := c.Param("minigameId")
	minigameId, err := strconv.Atoi(minigameChar)

	var response []leaderboard

	if err != nil {
		l.Errorf("error parsing minigameId %+v", err)
		return c.JSON(http.StatusInternalServerError, leaderBoardResponse{Leaderboard: response, Message: "error parsing minigame form url"})
	}

	db := config.DB

	var miniGame models.MiniGame

	if err := db.First(&miniGame, minigameId).Error; err != nil {
		l.Errorf("Error fetching miniGame from db %+v", err)
		return c.JSON(http.StatusBadRequest, leaderBoardResponse{Leaderboard: response, Message: "invalid minigame"})
	}

	query := "SELECT u.userName AS name, L.score AS score, u.department AS department FROM LeaderBoard AS L LEFT JOIN User AS u ON L.userId = u.id WHERE L.miniGameId = ? ORDER BY L.score DESC;"

	if err := db.Raw(query, minigameId).Scan(&response).Error; err != nil {
		l.Errorf("Error fetching leaderboard from db %+v", err)
		return c.JSON(http.StatusInternalServerError, leaderBoardResponse{Leaderboard: response, Message: "internal server error"})
	}

	l.Infof("getleaderBoard for minigame id %d is successful", minigameId)

	return c.JSON(http.StatusOK, leaderBoardResponse{Leaderboard: response, Message: "success"})
}
