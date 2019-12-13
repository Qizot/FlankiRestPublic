package models

import (
	"FlankiRest/errors"
	"encoding/json"
	"github.com/evanphx/json-patch"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"reflect"
	"strconv"
	"time"
)



type LobbyModel struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

type Lobby struct {
	LobbyModel
	OwnerID     uint      `json:"lobby_owner"`
	Name        string    `json:"name"`
	PlayerLimit uint      `json:"player_limit"`
	Password    string    `json:"password,omitempty"`
	Private     *bool     `json:"private"`
	Closed      *bool     `json:"closed"`
	Teams       []Team    `json:"teams"`
	Winner      TeamColor `json:"winner,omitempty"`
	Longitude   float64   `json:"longitude"`
	Latitude    float64   `json:"latitude"`
}

type LobbyListing struct {
	ID          uint      `json:"id"`
	OwnerID     uint      `json:"lobby_owner"`
	Name        string    `json:"name"`
	PlayerLimit uint      `json:"player_limit"`
	Private     bool      `json:"private"`
	Players     int       `json:"players"`
	CreatedAt   time.Time `json:"created_at"`
	Longitude   float64   `json:"longitude"`
	Latitude    float64   `json:"latitude"`
}

func NewLobbyListing(lobby *Lobby, db *gorm.DB) *LobbyListing {
	players := lobby.Teams[0].TeamEntriesCount(db) + lobby.Teams[1].TeamEntriesCount(db)
	return &LobbyListing{lobby.ID,lobby.OwnerID, lobby.Name, lobby.PlayerLimit, *lobby.Private, players, lobby.CreatedAt, lobby.Longitude, lobby.Latitude}
}



type UpdateLobby struct {
	Name 		string 	`json:"name,omitempty"`
	PlayerLimit uint   `json:"player_limit"`
	Password    string `json:"password,omitempty"`
	Private     *bool `json:"private"`
	Longitude   float64   `json:"longitude"`
	Latitude    float64   `json:"latitude"`
}

type MatchResult struct {
	ID        uint      `json:"lobby_id"`
	Winner    TeamColor `json:"winner"`
	Teams     *[]Team   `json:"teams"`
	Submitted time.Time `json:"finished"`
}

func (lobby *Lobby) IsEmpty() bool {
	return reflect.DeepEqual(UpdateAccount{}, lobby)
}


func (lobby *Lobby) Validate(db *gorm.DB) error {
	_, err := GetOwnersLobby(db, lobby.OwnerID, false)
	if err == nil {
		return errors.New("Player is already an owner of a lobby", 400)
	}

	if lobby.PlayerLimit < 4 || lobby.PlayerLimit > 20  {
		return errors.New("Wrong player limit, should be from 4 to 20 players", 400)
	}

	if size := len(lobby.Name); size < 4 || size > 50 {
		return errors.New("Wrong lobby's name size, should be from 4 to 50 characters", 400)
	}

	if lobby.Private == nil {
		private := false
		lobby.Private = &private
	}

	if size := len(lobby.Password); *lobby.Private == true && (size < 4 || size > 20) {
		return errors.New("Wrong password size, should be from 4 to 20 characters", 400)
	}
	return nil
}

func (lobby* Lobby) GetUpdateStruct() *UpdateLobby {
	updateLobby := &UpdateLobby{}
	updateLobby.PlayerLimit = lobby.PlayerLimit
	updateLobby.Private     = lobby.Private
	updateLobby.Password    = lobby.Password
	updateLobby.Name        = lobby.Name
	updateLobby.Longitude   = lobby.Longitude
	updateLobby.Latitude    = lobby.Latitude
	return updateLobby
}

func (lobby *Lobby) Create(db *gorm.DB) error {
	check := false
	err := lobby.Validate(db)
	if err != nil {
		return err
	}
	lobby.Closed = &check
	lobby.Winner = NoneTeam
	if *lobby.Private == true {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(lobby.Password), bcrypt.MinCost)
		lobby.Password = string(hashedPassword)
	} else {
		lobby.Password = ""
	}

	blueTeam := Team{TeamColor: Blue, TeamEntries: []TeamEntry{}}
	redTeam := Team{TeamColor: Red, TeamEntries: []TeamEntry{}}
	lobby.Teams = []Team{blueTeam, redTeam}

	err = db.Create(lobby).Error
	if err != nil {
		return errors.New("Failed to create lobby, connection error.", 500)
	}
	// Delete password
	lobby.Password = ""
	return nil
}

func (lobby *Lobby) Delete(db *gorm.DB) error {
	if *lobby.Closed == true {
		return errors.New("Cannot delete closed lobby", 400)
	}
	err := lobby.Close(db)
	if err != nil {
		return err
	}
	err = db.Unscoped().Delete(lobby).Error // player entries and teams are set to delete on CASCADE
	if err != nil {
		return errors.DatabaseError(err)
	}
	return nil
}


func (lobby *Lobby) Close(db *gorm.DB) error {
	if *lobby.Closed == true {
		return errors.New("Last lobby has been already closed", 400)
	}

	playersToBeStoppedPlaying, err  := lobby.GetLobbyPlayersIds(db)
	if err!= nil {
		return err
	}

	*lobby.Closed = true
	err = db.Save(lobby).Error
	if err != nil {
		return errors.New("Database error while closing lobby", 500)
	}

	for _, id := range playersToBeStoppedPlaying {
		err = db.Model(&Account{}).Where("id = ?", id).Update("playing", false).Error
		if err != nil {
			return errors.New("Database error while changing players' playing status: " + err.Error() , 500)
		}
	}
	return nil
}

// this lobby have to contain owner's id beside date to be updated
func (lobby *Lobby) Update(db *gorm.DB) error {

	ownersLobby, err := GetOwnersLobby(db, lobby.OwnerID, false)
	if err == gorm.ErrRecordNotFound {
		return errors.New("Player was not an owner of any ownersLobby", 400)
	}
	if err != nil {
		return errors.New("Database error while trying to get ownersLobby: " + err.Error(), 500)
	}

	// have to assign those fields because json probably assigned them defaults when they were not present in the request
	if lobby.PlayerLimit == 0 {
		lobby.PlayerLimit = ownersLobby.PlayerLimit
	}
	if lobby.Private == nil {
		lobby.Private = ownersLobby.Private
	}

	if  lobby.PlayerLimit < 4 || lobby.PlayerLimit > 20 {
		return errors.New("Wrong player limit, should be from 4 to 20 players", 400)
	}

	if *lobby.Private == true && lobby.Password == "" {
		return errors.New("Password was not given while ownersLobby is set to private", 400)
	}

	if size := len(lobby.Password); *lobby.Private == true && (size < 4 || size > 20) {
		return errors.New("Wrong password size, should be from 4 to 20 characters", 400)
	}

	if lobby.Name != "" {
		if size := len(lobby.Name); size < 4 || size > 50 {
			return errors.New("Wrong ownersLobby's name size, should be from 4 to 50 characters", 400)
		}
	}

	// such a hack, no one ever lived on either coordinate 0 I guess, not in Poland at least
	if lobby.Latitude == 0 {
		lobby.Latitude = ownersLobby.Latitude
	}
	if lobby.Longitude == 0 {
		lobby.Longitude = ownersLobby.Longitude
	}

	updateLobby := lobby.GetUpdateStruct()

	updateJson, _ := json.Marshal(updateLobby)
	lobbyJson, _ := json.Marshal(ownersLobby)
	newJson, _ := jsonpatch.MergePatch(lobbyJson, updateJson)

	newLobby := &Lobby{}
	err = json.Unmarshal(newJson, newLobby)
	if err != nil {
		return errors.New("Failed to update ownersLobby information", 500)
	}

	if *newLobby.Private == true {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newLobby.Password), bcrypt.DefaultCost)
		newLobby.Password = string(hashedPassword)
	} else {
		newLobby.Password = ""
	}
	err = db.Save(newLobby).Error
	if err != nil {
		return errors.New("Database error: " + err.Error(), 500)
	}
	return nil
}

func (lobby *Lobby) GetLobbyPlayersIds(db *gorm.DB) ([]uint, error) {
	var teams []Team
	err := db.Model(&lobby).Related(&teams).Error
	if err != nil {
		return nil, errors.New("Cannot find lobby's teams: " + err.Error(), 500)
	}
	if len(teams) != 2 {
		return nil, errors.New("Fetched only " + strconv.Itoa(len(teams)) + " teams but should be 2", 500)
	}

	var blueEntries []TeamEntry
	var redEntries []TeamEntry
	err = db.Model(&teams[0]).Related(&blueEntries).Error
	err = db.Model(&teams[1]).Related(&redEntries).Error
	if err != nil {
		return nil, errors.New("Error while fetching players from team: " + err.Error(), 500)
	}
	ids := make([]uint, 0, len(blueEntries)+len(redEntries))
	for _, id := range blueEntries {
		ids = append(ids, id.PlayerID)
	}
	for _, id := range redEntries {
		ids = append(ids, id.PlayerID)
	}
	return ids, nil
}
func GetOwnersLobby(db *gorm.DB, id uint, preloadTeamEntries bool) (*Lobby, error) {

	lobby := &Lobby{}
	var dbcon *gorm.DB
	dbcon = db
	if preloadTeamEntries {
		dbcon = db.Preload("Teams.TeamEntries")
	}
	err := dbcon.Where("owner_id = ? AND closed = ?", id, false).First(&lobby).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("Player is not an owner of any opened lobby", 404)
		}
		return nil, errors.New("Error while getting owner's lobby: " + err.Error(),500)
	}
	return lobby, nil
}

func GetLobbyByIdFunc(db *gorm.DB, id uint) (*Lobby, error) {
	lobby := &Lobby{}
	err := db.Preload("Teams.TeamEntries").First(&lobby, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("Lobby has not been found", 404)
		}
		return nil, errors.New("Database error while getting lobby by id: " + err.Error(), 500)
	}
	return lobby, nil
}

func GetAllLobbies(db *gorm.DB, closed bool) ([]*Lobby, error) {
	var lobbies []*Lobby
	err := db.Preload("Teams.TeamEntries").Where("closed = ?", closed).Order("updated_at desc").Limit(100).Find(&lobbies).Error

	if err != nil {
		return nil, errors.New("Database error while getting list of lobbies: " + err.Error(), 500)
	}
	for _, lobby := range lobbies {
		lobby.Password = ""
	}
	return lobbies, nil
}

