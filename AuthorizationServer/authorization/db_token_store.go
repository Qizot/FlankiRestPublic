package authorization

import (
	"AuthorizationServer/database"
	"encoding/json"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sirupsen/logrus"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
	"time"
)




type StoreItemModel struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	//DeletedAt *time.Time `sql:"index"`
}


// StoreItem data item
type StoreItem struct {
	StoreItemModel
	ExpiredAt int64
	UserID    string
	Code      string `gorm:"type:varchar(512)"`
	Access    string `gorm:"type:varchar(512)"`
	Refresh   string `gorm:"type:varchar(512)"`
	Data      string `gorm:"type:text"`
}

func NewStoreWithDB(apidb *database.AuthDatabase, gcInterval int, dbInterval int, logger *logrus.Logger) (store *Store, err error) {

	store = &Store{
		db:        apidb,
		tableName: "oauth2_token",
		logEntry: logger.WithField("prefix", "[TOKEN STORE]"),
	}

	// garbage collecting old tokens
	if gcInterval < 0 {
		gcInterval = 600
	}

	// interval every each db connection will be checked and when needed reconnected
	if dbInterval < 0 {
		dbInterval = 60
	}
	store.gcTicker = time.NewTicker(time.Second * time.Duration(gcInterval))
	store.dbTicker = time.NewTicker(time.Second * time.Duration(dbInterval))

	go store.databaseCheck()
	go store.gc()
	return
}

// Store mysql token store
type Store struct {
	tableName string
	db        *database.AuthDatabase
	gcTicker  *time.Ticker
	dbTicker  *time.Ticker
	logEntry  *logrus.Entry
}


func (s *Store) Close() {
	s.gcTicker.Stop()
	s.dbTicker.Stop()
}

// when database is down and then reconnects, token model migration might be needed
// here comes this function which will be checking every interval for datatable existence when database is up again
// token store should not try to reconnect with database because it should be done by auth server
func (s *Store) databaseCheck() {

	for range s.dbTicker.C {
		db := s.db.DB()
		if err := db.DB().Ping(); err != nil {
			s.logEntry.Error("Database connection is closed")
			continue
		}
		if !db.HasTable(s.tableName) {
			if err := db.Table(s.tableName).CreateTable(&StoreItem{}).Error; err != nil {
				s.logEntry.Error(err.Error())
			} else {
				s.logEntry.Info("Created database schema")
			}
		}
	}
}

// garbage collector for expired tokens
func (s *Store) gc() {
	for range s.gcTicker.C {
		db := s.db.DB()
		now := time.Now().Unix()
		var count int
		if err := db.Table(s.tableName).Where("expired_at < ?", now).Count(&count).Error; err != nil {
			s.logEntry.Error(err.Error())
		}
		if count > 0 {
			if err := db.Table(s.tableName).Where("expired_at < ?", now).Delete(&StoreItem{}).Error; err != nil {
				s.logEntry.Error(err.Error())
			}
		}
	}
}

// Create create and store the new token information
func (s *Store) Create(info oauth2.TokenInfo) error {
	db := s.db.DB()
	jv, err := json.Marshal(info)
	if err != nil {
		return err
	}
	item := &StoreItem{
		Data: string(jv),
		UserID: info.GetUserID(),
	}

	if code := info.GetCode(); code != "" {
		item.Code = code
		item.ExpiredAt = info.GetCodeCreateAt().Add(info.GetCodeExpiresIn()).Unix()
	} else {
		item.Access = info.GetAccess()
		item.ExpiredAt = info.GetAccessCreateAt().Add(info.GetAccessExpiresIn()).Unix()

		if refresh := info.GetRefresh(); refresh != "" {
			item.Refresh = info.GetRefresh()
			item.ExpiredAt = info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn()).Unix()
		}
	}
	err = db.Table(s.tableName).Where("user_id = ?", info.GetUserID()).Delete(StoreItem{}).Error

	return db.Table(s.tableName).Create(item).Error
}

// RemoveByCode delete the authorization code
func (s *Store) RemoveByCode(code string) error {
	return s.db.DB().Table(s.tableName).Where("code = ?", code).Update("code", "").Error
}

// RemoveByAccess use the access token to delete the token information
func (s *Store) RemoveByAccess(access string) error {
	return s.db.DB().Table(s.tableName).Where("access = ?", access).Update("access", "").Error
}

// RemoveByRefresh use the refresh token to delete the token information
func (s *Store) RemoveByRefresh(refresh string) error {
	return s.db.DB().Table(s.tableName).Where("refresh = ?", refresh).Update("refresh", "").Error
}

func (s *Store) toTokenInfo(data string) oauth2.TokenInfo {
	var tm models.Token
	err := json.Unmarshal([]byte(data), &tm)
	if err != nil {
		return nil
	}
	return &tm
}

// GetByCode use the authorization code for token information data
func (s *Store) GetByCode(code string) (oauth2.TokenInfo, error) {
	if code == "" {
		return nil, nil
	}

	var item StoreItem
	if err := s.db.DB().Table(s.tableName).Where("code = ?", code).Find(&item).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return s.toTokenInfo(item.Data), nil
}



// GetByAccess use the access token for token information data
func (s *Store) GetByAccess(access string) (oauth2.TokenInfo, error) {
	if access == "" {
		return nil, nil
	}
	var item StoreItem
	if err := s.db.DB().Table(s.tableName).Where("access = ?", access).Find(&item).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrTokenNotFound
		}
		return nil, ErrDatabaseError
	}
	return s.toTokenInfo(item.Data), nil
}

// GetByRefresh use the refresh token for token information data
func (s *Store) GetByRefresh(refresh string) (oauth2.TokenInfo, error) {
	if refresh == "" {
		return nil, nil
	}

	var item StoreItem
	if err := s.db.DB().Table(s.tableName).Where("refresh = ?", refresh).Find(&item).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return s.toTokenInfo(item.Data), nil
}