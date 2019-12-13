package models

import (
	"FlankiRest/errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type TeamEntry struct {
	ID uint `gorm:"primary_key" json:"-"`
	PlayerID uint `json:"player_id"`
	Nickname string `json:"nickname,omitempty"`
	TeamID uint `json:"-"`
}

type TeamColor string

const (
	Blue TeamColor = "blue"
	Red TeamColor = "red"
	NoneTeam TeamColor = ""
	Spectator TeamColor = "spectator"
)

type Team struct {
	ID uint `gorm:"primary_key" json:"id"`
	TeamEntries []TeamEntry `json:"players"`
	TeamColor TeamColor `json:"team_color"`
	LobbyID uint `json:"-"`
}

type LobbyRequest struct {
	TeamColor TeamColor `json:"team_color"`
	Password string `json:"password"`
}

func (request *LobbyRequest) Validate() error {
	if request.TeamColor != Red && request.TeamColor != Blue {
		return errors.New("Invalid team color", 400)
	}
	return nil
}

func (request *LobbyRequest) CheckPassword(lobby *Lobby) bool {
	err := bcrypt.CompareHashAndPassword([]byte(lobby.Password), []byte(request.Password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return false
	}
	return true
}

func (team *Team) AddNewEntry(db *gorm.DB, entry * TeamEntry) error {
	return db.Model(&team).Association("TeamEntries").Append(entry).Error
}

func (team *Team) ContainsPlayerWithId(db *gorm.DB, id uint) bool {
	var count []uint
	err := db.Raw(fmt.Sprintf("SELECT count(team_entries.player_id) FROM teams INNER JOIN team_entries ON teams.id = team_entries.team_id WHERE teams.id = %d AND team_entries.player_id = %d", team.ID, id)).Pluck("id",&count).Error
	if err != nil {
		return false
	}
	return count[0] > 0
}

func (team *Team) DeleteEntryWithPlayerId(db *gorm.DB, id uint) error {
	var entries []TeamEntry
	err := db.Model(&team).Association("TeamEntries").Find(&entries).Error
	if err != nil {
		return errors.New("Database error while fetching team's players: " + err.Error(), 500)
	}
	for _, entry := range entries {
		if entry.PlayerID == id {
			err = db.Delete(&entry).Error
			if err != nil {
				return errors.New("Database error while deleting entry from team: " + err.Error(), 500)
			}
			err = db.Model(&Account{}).Where("id = ?", id).Update("playing", false).Error
			if err != nil {
				return errors.New("Database error while updating player's playing status: " + err.Error(), 500)
			}
			return nil
		}
	}
	return errors.New("Player as not been found in a team", 404)
}

func (team *Team) TeamEntriesCount(db *gorm.DB) int {
	return db.Model(&team).Association("TeamEntries").Count()
}


