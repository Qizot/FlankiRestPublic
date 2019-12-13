package models

import (
	"FlankiRest/errors"
	"github.com/jinzhu/gorm"
)

type Player struct {
	ID 			uint 	`json:"id"`
	Nickname    string `json:"nickname"`
	Sex         string `json:"sex"`
	Description string `json:"description"`
	Playing     bool   `json:"playing"`
}

func GetAllPlayersFunc(db *gorm.DB) ([]*Player, error) {
	var account Account
	var players []*Player
	err := db.Model(&account).Select("id, nickname, sex, description, playing").Scan(&players).Error

	if err != nil {
		return players, errors.New("Database error while fetching list of players" + err.Error(), 500)
	}
	return players, nil
}

func GetPlayerByIdFunc(db *gorm.DB, id uint) (*Player, error) {
	//var account Account
	var players []*Player
	err := db.Model(&Account{}).Select("id, nickname, sex, description, playing").Where("id = ?", id).Scan(&players).Error
	if err != nil {
		return nil, errors.New("Database error while fetching player by id" + err.Error(), 500)
	}
	if len(players) == 0 {
		return nil, errors.New("Player has not been found", 404)
	}
	return players[0], nil
}


