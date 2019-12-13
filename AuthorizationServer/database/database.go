package database



import "github.com/jinzhu/gorm"

// this struct is just a dummy wrapper for db connection pointer
// its main goal is to keep db pointer for reconnecting goroutine
type AuthDatabase struct {
	db *gorm.DB
}

func (apidb *AuthDatabase) DB() *gorm.DB {
	return apidb.db
}

func (apidb *AuthDatabase) NewConnection(uri string) (err error) {
	apidb.db, err = gorm.Open("postgres", uri)
	return
}

func (apidb *AuthDatabase) Close() error {
	return apidb.db.Close()
}


