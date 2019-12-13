package controllers

import (
	"FlankiRest/database"
	"FlankiRest/errors"
	"FlankiRest/models"
	u "FlankiRest/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type StatisticsController struct {
	DB *database.ApiDatabase
	logger *logrus.Logger
}

func NewStatisticsController(db *database.ApiDatabase, logger *logrus.Logger) *StatisticsController {
	return &StatisticsController{db, logger}
}

func (controller *StatisticsController) Logger() *logrus.Logger {
	return controller.logger
}

func (controller *StatisticsController) GetPlayerSummary(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	vars := mux.Vars(r)
	playerID, err := strconv.Atoi(vars["id"])
	if err != nil {
		u.ApiErrorResponse(w, errors.New("Invalid player id", 400))
		return
	}
	summary, err := models.GetPlayersSummary(db, uint(playerID))
	if err != nil {
		u.ApiErrorResponse(w, err)
	}
	u.SimpleRespond(w, summary)
	return
}

func (controller *StatisticsController) GetPlayersRanking(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	ranking, err := models.GetPlayersRanking(db)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	u.SimpleRespond(w, ranking)
	return
}