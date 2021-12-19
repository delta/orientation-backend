package leaderboard

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	logger "github.com/sirupsen/logrus"
	"strconv"
)

func RegisterRoutes(v *echo.Group) {
	v.POST("/addscore", AddScore)
}

type Score struct {
	Game  string `json:"game"`
	Score string `json:"score"`
}

func AddScore(c echo.Context) error {
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
