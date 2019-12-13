package utils

import (
	"AuthorizationServer/database"
	"time"
)

type DatabseReconnectTicker struct {
	db *database.AuthDatabase
	uri string
	Ticker *time.Ticker
}

func NewDatabaseReconnectTicker(conn *database.AuthDatabase,uri string,  interval time.Duration) *DatabseReconnectTicker {
	drt := &DatabseReconnectTicker{conn,uri, time.NewTicker(time.Second * interval)}
	return drt
}

func (drt *DatabseReconnectTicker) TryReconnect() (bool, error) {
	if err := drt.db.DB().DB().Ping(); err != nil {
		err = drt.db.NewConnection(drt.uri)
		
		if err != nil {
			return false, err
		} else {
			return true, err
		}
	}
	return false, nil
}