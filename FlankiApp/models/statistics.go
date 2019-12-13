package models

import (
	"FlankiRest/errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"math"
)

type PlayerStatisticsEntry struct {
	ID       uint   `json:"id"`
	PlayerID uint   `json:"player_id"`
	LobbyId  uint   `json:"lobby_id"`
	Win      bool   `json:"is_win"`
	Points   int    `json:"points"`
}

type PlayerSummary struct {
	PlayerID uint   `json:"player_id"`
	Nickname string `json:"nickname,omitempty"`
	Points   int    `json:"points"`
	Wins     int    `json:"wins"`
	Loses    int    `json:"loses"`
}

func (summary *PlayerSummary) GetQuickSummary() *QuickSummary {
	return &QuickSummary{summary.Points, summary.Wins, summary.Loses}
}

type QuickSummary struct {
	Points   int    `json:"points"`
	Wins     int    `json:"wins"`
	Loses    int    `json:"loses"`
}



func GetPlayersSummary(db *gorm.DB, id uint) (*PlayerSummary, error) {
	var summary PlayerSummary
	var account Account
	err := db.Model(&Account{}).Select("id, nickname").Where("id = ?", id).First(&account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.PlayerNotFound
		}
		return nil, errors.New("Error while checking players ranking", 500)
	}
	summary.Nickname = account.Nickname

	querry := fmt.Sprintf(`Select %d as player_id,
	(select count(id) from player_statistics_entries where player_id = %d and win = 'true') as wins,
		(select count(id) from player_statistics_entries where player_id = %d and  win = 'false') as loses,
		(select sum(points) from player_statistics_entries where player_id = %d ) as points;`, id, id, id, id)
	err = db.Raw(querry).Scan(&summary).Error
	if err != nil {
		return nil, errors.New("Couldn't get players summary: " + err.Error(), 500)
	}
	return &summary, nil
}

func GetPlayersRanking(db *gorm.DB) ([]PlayerSummary, error) {
	var  summaries []PlayerSummary
	querry := `select player_id, (select nickname from accounts where id = player_id),
  					  	sum(case when win = 'true' then 1 else 0 end) as wins, 
						sum(case when win = 'false' then 1 else 0 end) as loses, 
						sum(points) as points  from player_statistics_entries 
						group by player_id
						order by points desc`
	err := db.Raw(querry).Scan(&summaries).Error
	if err != nil {
		return nil, errors.New("Couldn't get players ranking: " + err.Error(), 500)
	}
	return summaries, nil
}


// Points are computed based on number of players in your team, if  your team wins
// you gain 10 points for each player in your team, but if you lose you lose 10 points for each teammate

func ComputePoints(db *gorm.DB, lobby *Lobby, currentTeam TeamColor) (int, error) {

	const multiplier = 1.5
	// penalty is zero for now
	const penalty = 0
	winner := lobby.Winner
	var teams []Team
	err := db.Model(&lobby).Related(&teams).Error
	if err != nil {
		return 0, errors.New("Encountered error while fetching teams from lobby", 500)
	}
	for _, team := range teams {

		// it means that we caught enemy team, we calculate points based on the number of enemies
		if team.TeamColor != currentTeam {
			players := team.TeamEntriesCount(db)

			// if the number of enemy players equals 0 we then apply the penalty score
			// multiplier ^ 0 equals 1 so if there were no players then substracting more than
			// 1 (floating points arithmetics) we should get negative score which will indicate to apply penalty
			points := (math.Pow(multiplier, float64(players))-1) * 10

			//enemy team had no players so apply penalty for the winners
			if currentTeam == winner && points <= 0 {
				points = penalty
			}

			// if lost, change points to negative and scale down
			if lobby.Winner != currentTeam {
				points *= -0.5
			}
			return int(math.Floor(points)), nil
		}
	}
	return 0,nil
}

func SubmitMatch(db *gorm.DB, lobby *Lobby) error {
	var teams []Team
	err := db.Model(&lobby).Related(&teams).Error
	if err != nil {
		return errors.New("Encountered error while fetching teams from lobby", 500)
	}
	for _, team := range teams {

		points, err := ComputePoints(db, lobby, team.TeamColor)
		if err != nil {
			return err
		}

		var won bool
		switch lobby.Winner {
		case team.TeamColor:
			won = true
		default:
			won = false
		}

		var players []TeamEntry
		err = db.Model(&team).Related(&players).Error
		if err != nil {
			return errors.New("Encountered error while fetching teams entries from team", 500)
		}
		for _, entry := range players {
			res := &PlayerStatisticsEntry{PlayerID: entry.PlayerID, LobbyId: lobby.ID, Points: points, Win: won}
			err = db.Create(res).Error
			if err != nil {
				return errors.New("Error while saving results"+err.Error(), 500)
			}
		}
	}
	return nil
}