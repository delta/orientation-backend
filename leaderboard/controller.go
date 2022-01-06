package leaderboard

import (
	"net/http"

	"github.com/delta/orientation-backend/auth"
	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/models"
	"github.com/labstack/echo/v4"
	logger "github.com/sirupsen/logrus"
)

func handleAddScore(c echo.Context, game_name string, score int) error {
	l := logger.WithFields(logger.Fields{
		"method": "leaderboard/controller/handleAddScore",
	})

	user, err := auth.GetCurrentUser(c)

	l.Errorf("user not found %+v", err)

	if err != nil {
		return err
	}

	var game models.MiniGame

	err = config.DB.Where("name = ?", game_name).First(&game).Error

	if err != nil {
		l.Errorf("minigame not found")
		return err
	}

	var leader models.LeaderBoard

	err = config.DB.Where("miniGameId = ? AND userid = ?", game.ID, user.ID).First(&leader).Error

	if err != nil {
		l.Infof("Couldn't find user in leaderboard for given name. Creating new record")
		record := models.LeaderBoard{GameId: game.ID, UserId: user.ID, Score: score}
		if err := config.DB.Create(&record).Error; err != nil {
			return err
		}

		return nil
	}

	if leader.Score < score {
		l.Infof("New score is greater. Updating")
		leader.Score = score
		if err := config.DB.Save(&leader).Error; err != nil {
			return err
		}
	}

	return nil
}

func getLeaderBoard(c echo.Context) error {
	l := logger.WithFields(logger.Fields{
		"method": "leaderboard/routes/getLeaderBoard",
		"name":   c.Param("minigameName"),
	})

	l.Infoln("getleaderBoard requested")

	minigameName := c.Param("minigameName")

	var response []leaderboard

	db := config.DB

	var miniGame models.MiniGame

	if err := db.Where("name = ?", minigameName).First(&miniGame).Error; err != nil {
		l.Errorf("Error fetching miniGame from db %+v", err)
		return c.JSON(http.StatusBadRequest, leaderBoardResponse{Leaderboard: response, Message: "invalid minigame"})
	}

	query := "SELECT u.userName AS username, L.score AS score, u.name as name, u.department AS department, u.id as id, u.id FROM LeaderBoard AS L LEFT JOIN User AS u ON L.userId = u.id WHERE L.miniGameId = ? ORDER BY L.score DESC;"

	if err := db.Raw(query, miniGame.ID).Scan(&response).Error; err != nil {
		l.Errorf("Error fetching leaderboard from db %+v", err)
		return c.JSON(http.StatusInternalServerError, leaderBoardResponse{Leaderboard: response, Message: "internal server error"})
	}

	l.Infof("getleaderBoard for minigame id %s is successful", minigameName)

	return c.JSON(http.StatusOK, leaderBoardResponse{Leaderboard: response, Message: "success"})
}
