package leaderboard

import (
	"net/http"

	"encoding/base64"
	"github.com/labstack/echo/v4"
	logger "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

func RegisterRoutes(v *echo.Group) {
	v.POST("/addscore", AddScore)
}

type Score struct {
	Data string `json:"data"`
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
	dec, _ := base64.StdEncoding.DecodeString(s.Data)
	split := strings.Split(string(dec), "$")
	if len(split) < 3 {
		l.Errorf("Invalid data. Not enough values")
		return c.JSON(http.StatusBadRequest, ScoreAddStatus{Status: false, Message: "Couldn't add score"})
	}
	game_name := split[1]
	sc, err := strconv.Atoi(split[2])
	if err != nil {
		l.Errorf("Score is not an integer")
		return c.JSON(http.StatusBadRequest, ScoreAddStatus{Status: false, Message: "Couldn't add score"})
	}
	err = handleAddScore(c, game_name, sc)
	if err != nil {

		l.Errorf("Couldn't add score")
		return c.JSON(http.StatusBadRequest, ScoreAddStatus{Status: false, Message: "Couldn't add score"})
	}
	return c.JSON(http.StatusOK, ScoreAddStatus{Status: true, Message: "Added successfully"})
}
