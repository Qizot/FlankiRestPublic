package controllers

import (
	"FlankiRest/database"
	"FlankiRest/errors"
	"FlankiRest/models"
	u "FlankiRest/utils"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type LobbyController struct {
	DB *database.ApiDatabase
	logger *logrus.Logger
}

func NewLobbyController(db *database.ApiDatabase, logger *logrus.Logger) *LobbyController {
	return &LobbyController{db, logger}
}

func (controller *LobbyController) Logger() *logrus.Logger {
	return controller.logger
}

func (controller *LobbyController) CreateLobby(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	lobby := &models.Lobby{}
	err := json.NewDecoder(r.Body).Decode(&lobby)

	if err != nil || lobby.IsEmpty() {
		u.ApiErrorResponse(w, errors.BadJsonRequestFormat)
		return
	}
	id, err := u.GetUserIdFromContext(r.Context())
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	lobby.OwnerID = id

	err = lobby.Create(db)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	lobby.Password = ""
	u.SimpleRespond(w, lobby)
	return
}

func (controller *LobbyController) DeleteLobby(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()

	id, err := u.GetUserIdFromContext(r.Context())
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}

	lobby, err := models.GetOwnersLobby(db, id, false)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}

	err = lobby.Delete(db)
	if err != nil {
		u.ApiErrorResponse(w, err)
	}

	u.SimpleRespond(w, u.TextMessage("Lobby has been deleted"))
	return
}

func (controller *LobbyController) UpdateLobby(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	lobbyChange := &models.Lobby{}
	err := json.NewDecoder(r.Body).Decode(lobbyChange)
	if err != nil || lobbyChange.IsEmpty() {
		u.ApiErrorResponse(w, errors.BadJsonRequestFormat)
		return
	}
	id, err := u.GetUserIdFromContext(r.Context())
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	lobbyChange.OwnerID = id

	err = lobbyChange.Update(db)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	u.SimpleRespond(w, u.TextMessage("Lobby has been updated"))
	return
}

func (controller *LobbyController) CloseLobby(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	id, err := u.GetUserIdFromContext(r.Context())
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	lobby, err := models.GetOwnersLobby(db, id, false)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	err = lobby.Close(db)
	if err != nil {
		u.ApiErrorResponse(w,err)
		return
	}
	u.SimpleRespond(w, u.TextMessage("Lobby has been closed"))
}

func (controller *LobbyController) GetLobbyById(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	lobby, err := models.GetLobbyByIdFunc(db, uint(id))
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	lobby.Password = ""
	u.SimpleRespond(w, lobby)
	return
}

func (controller *LobbyController) GetAllLobbies(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	// gets list of all opened lobbies
	lobbies, err := models.GetAllLobbies(db, false)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	lobbiesListing := []*models.LobbyListing{}
	for _, l := range lobbies {
		lobbiesListing = append(lobbiesListing, models.NewLobbyListing(l,db))
	}
	u.SimpleRespond(w, lobbiesListing)
	return
}

func (controller *LobbyController) JoinLobbyTeam(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	vars := mux.Vars(r)
	lobbyID, _ := strconv.Atoi(vars["id"])
	playerID, err := u.GetUserIdFromContext(r.Context())
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	joinRequest := &models.LobbyRequest{}


	err = json.NewDecoder(r.Body).Decode(joinRequest)
	if err != nil || *joinRequest == (models.LobbyRequest{}) {
		u.ApiErrorResponse(w, errors.BadJsonRequestFormat)
		return
	}
	if joinRequest.TeamColor == models.Spectator {
		lobby, err := models.GetLobbyByIdFunc(db, uint(lobbyID))
		if err != nil {
			u.ApiErrorResponse(w, err)
			return
		}
		u.SimpleRespond(w,lobby)
		return
	}

	account, err := models.GetAccountById(db, playerID)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	if account.Playing == true {
		u.ApiErrorResponse(w, errors.New("User is already playing, can't join to new team", 400))
		return
	}

	err = joinRequest.Validate()
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}

	lobby, err := models.GetLobbyByIdFunc(db, uint(lobbyID))
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}

	if *lobby.Private == true {
		if ok := joinRequest.CheckPassword(lobby); !ok {
			u.ApiErrorResponse(w, errors.UnauthorizedLobbyJoinRequest)
			return
		}
	}

	var teams []models.Team
	err = db.Model(&lobby).Association("Teams").Find(&teams).Error
	if err != nil {
		u.ApiErrorResponse(w, errors.New("Database error while trying extract teams' data from lobby: " + err.Error(), 500))
		return
	}

	blueTeam := teams[0]
	redTeam := teams[1]

	if blueTeam.TeamEntriesCount(db) + redTeam.TeamEntriesCount(db) >= int(lobby.PlayerLimit) {
		u.ApiErrorResponse(w, errors.LobbyIsFull)
		return
	}
	var team *models.Team
	if joinRequest.TeamColor == models.Blue {
		team = &blueTeam
	}
	if joinRequest.TeamColor == models.Red {
		team = &redTeam
	}
	player := &models.Account{}
	err = db.Model(&models.Account{}).Where("id = ?",playerID).Select("nickname").First(&player).Error
	if err != nil {
		u.ApiErrorResponse(w, errors.DatabaseError(err))
		return
	}
	entry := &models.TeamEntry{PlayerID: playerID,Nickname: player.Nickname}
	err = team.AddNewEntry(db, entry)
	if err != nil {
		u.ApiErrorResponse(w, errors.New("Database error while joining lobby: " + err.Error(), 500))
		return
	}
	err = db.Model(&account).Update("playing", true).Error
	if err != nil {
		u.ApiErrorResponse(w, errors.DatabaseError(err))
		return
	}

	u.SimpleRespond(w, u.TextMessage("Joining to the lobby has been successful!"))
	return
}

func (controller *LobbyController) LeaveLobby(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()

	lobby := &models.Lobby{}
	playerID, err := u.GetUserIdFromContext(r.Context())
	player, err := models.GetAccountById(db, playerID)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	err = db.Joins("JOIN teams on lobbies.id = teams.lobby_id").
		Joins("JOIN team_entries on teams.id = team_entries.team_id").
		Where("team_entries.player_id = ? and lobbies.closed = false", playerID).
		First(&lobby).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// encountered bug where all lobbies were closed but player playing status was still true
			// pray that it will happen again to solve this problem
			if player.Playing == true {
				controller.Logger().WithField("prefix", "[BUG]").Error("Player is still playing but is not present in any opened lobby")
				player.Playing = false
				err = db.Save(player).Error
				if err != nil {
					err = errors.DatabaseError(err)
				}
			} else {
				err = errors.PlayerNotActive
			}
		} else {
			err =  errors.DatabaseError(err)
		}
		u.ApiErrorResponse(w,err)
		return
	}

	var teams []models.Team
	err = db.Model(&lobby).Association("Teams").Find(&teams).Error
	if err != nil {
		u.ApiErrorResponse(w, errors.DatabaseError(err))
		return
	}
	for _, team := range teams {
		if team.ContainsPlayerWithId(db, playerID) {
			err := team.DeleteEntryWithPlayerId(db, playerID)
			if err != nil {
				u.ApiErrorResponse(w, err)
				return
			}
			u.SimpleRespond(w, u.TextMessage("Left the lobby"))
			return
		}
	}
	u.ApiErrorResponse(w, errors.PlayerNotFoundInAnyTeam)
	return
}

func (controller *LobbyController) KickPlayerFromLobby(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()

	ownerID, err := u.GetUserIdFromContext(r.Context())
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}

	lobby, err := models.GetOwnersLobby(db, ownerID, true)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}

	type ID struct {
		ID uint `json:"player_id"`
	}
	playerID := &ID{}
	err = json.NewDecoder(r.Body).Decode(playerID)

	if err != nil || playerID.ID == 0 {
		u.ApiErrorResponse(w, errors.BadJsonRequestFormat)
		return
	}

	for _, team := range lobby.Teams {
		err = team.DeleteEntryWithPlayerId(db, playerID.ID)
		if err == nil {
			u.SimpleRespond(w, u.TextMessage("Player has been kicked out of lobby"))
			return
		}
	}
	if err == gorm.ErrRecordNotFound {
		u.ApiErrorResponse(w, errors.PlayerNotFoundInAnyTeam)
		return
	}
	u.ApiErrorResponse(w, errors.DatabaseError(err))
	return
}

// submits lobby results and end the game
func (controller *LobbyController) SubmitResults(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	ownerID, err := u.GetUserIdFromContext(r.Context())
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	lobby, err := models.GetOwnersLobby(db, ownerID, false)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	type Winner struct {
		TeamWin models.TeamColor `json:"winner"`
	}
	winner := &Winner{}
	err = json.NewDecoder(r.Body).Decode(winner)
	if err != nil || winner.TeamWin == "" {
		u.ApiErrorResponse(w, errors.BadJsonRequestFormat)
		return
	}
	if winner.TeamWin != models.Blue && winner.TeamWin != models.Red {
		u.ApiErrorResponse(w, errors.New("Invalid winner color, should be either blue or red",400))
		return
	}

	lobby.Winner = winner.TeamWin

	err = models.SubmitMatch(db, lobby)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}

	err = lobby.Close(db)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}

	err = db.Save(lobby).Error
	if err != nil {
		u.ApiErrorResponse(w, errors.DatabaseError(err))
		return
	}
	u.SimpleRespond(w, u.TextMessage("Results have been submitted"))
	return
}

// responds with list of closed lobbies which are said to be finished
func (controller *LobbyController) Results(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	var lobbies []*models.Lobby
	lobbies, err := models.GetAllLobbies(db, true)

	if err != nil {
		u.ApiErrorResponse(w, errors.DatabaseError(err))
		return
	}

	results := []*models.MatchResult{} // it's like this to marshal empty slice to '[]' and not to 'null'
	for _, lobby := range lobbies {
		results = append(results, &models.MatchResult{ID: lobby.ID, Winner: lobby.Winner, Teams: &lobby.Teams, Submitted: lobby.UpdatedAt})
	}

	u.SimpleRespond(w, results)
	return
}

// responds  with player's owner lobby
func (controller *LobbyController) OwnerLobby(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	ownerID, err := u.GetUserIdFromContext(r.Context())
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	lobby, err := models.GetOwnersLobby(db, ownerID, true)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	lobby.Password = ""
	u.SimpleRespond(w, lobby)
	return
}

// responds with player's current lobby if he is a member of one
func (controller *LobbyController) GetCurrentLobby(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	playerID, err := u.GetUserIdFromContext(r.Context())
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	lobby := &models.Lobby{}

	err = db.Joins("JOIN teams on lobbies.id = teams.lobby_id").
		Joins("JOIN team_entries on teams.id = team_entries.team_id").
		Preload("Teams.TeamEntries").
		Where("team_entries.player_id = ? and lobbies.closed = false", playerID).
		First(lobby).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			u.ApiErrorResponse(w, errors.PlayerNotActive)
			return
		}
		u.ApiErrorResponse(w, errors.DatabaseError(err))
		return
	}
	lobby.Password = ""
	u.SimpleRespond(w, lobby)
	return
}
