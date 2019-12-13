package controllers

import (
	"FlankiRest/database"
	"FlankiRest/models"
	u "FlankiRest/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type PlayerController struct {
	DB *database.ApiDatabase
	logger *logrus.Logger
}

func NewPlayerController(db *database.ApiDatabase, logger *logrus.Logger) *PlayerController {
	return &PlayerController{db, logger}
}

func (controller *PlayerController) Logger() *logrus.Logger {
	return controller.logger
}

func (controller *PlayerController) GetAllPlayers(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	players, err := models.GetAllPlayersFunc(db)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	u.SimpleRespond(w, players)
	return
}

func (controller *PlayerController) GetPlayerById(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	player, err := models.GetPlayerByIdFunc(db, uint(id))
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	response := map[string] interface {} {}
	response["player"] = player
	summary, err := models.GetPlayersSummary(db, player.ID)
	if summary != nil {
		response["summary"] = summary.GetQuickSummary()
	} else {
		response["summary"] = nil
	}
	u.SimpleRespond(w, response)
	return
}

