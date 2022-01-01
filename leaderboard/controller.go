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
	userCookie, err := c.Cookie(auth.CurrentConfig.Cookie.User.Name)

	l := logger.WithFields(logger.Fields{
		"method": "leaderboard/controller/handleAddScore",
	})
	l.Infof("Getting user cookie")
	if err != nil {
		l.Errorf("Couldn't find user cookie")
		return err
	}
	user, err := auth.Get_info_from_cookie(userCookie, "user")
	if err != nil {
		l.Errorf("Couldn't get user")
		return err
	}
	var game models.MiniGame
	err = config.DB.Where("name = ?", game_name).First(&game).Error
	if err != nil {
		l.Errorf("No mini game with the given name")
		return err
	}
	var leader models.LeaderBoard
	err = config.DB.Where("miniGameId = ? AND userid = ?", game.ID, user.Get_id()).First(&leader).Error
	if err != nil {
		l.Infof("Couldn't find user in leaderboard for given name. Creating new record")
		record := models.LeaderBoard{GameId: game.ID, UserId: user.Get_id(), Score: score}
		config.DB.Create(&record)
	} else {
		if leader.Score < score {
			l.Infof("New score is greater. Updating")
			leader.Score = score
			config.DB.Save(&leader)
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
