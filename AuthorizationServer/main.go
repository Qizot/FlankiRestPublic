package main

import (
	auth "AuthorizationServer/authorization"
	"AuthorizationServer/database"
	"AuthorizationServer/utils"
	"fmt"
	"os"
)

func main() {
	cfg := auth.GetAuthServerConfig()
	log := utils.AuthLoggerEntry()
	utils.SetEnvDebug()
	db := &database.AuthDatabase{}
	defer db.Close()
	dbUri := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", cfg.DBHost, cfg.DBUsername, cfg.DBName, cfg.DBPassword)
	dbTicker := utils.NewDatabaseReconnectTicker(db, dbUri, 10)

	dbSetup := func() {
		db.DB().SetLogger(&utils.GormLogger{})
		if debug := os.Getenv("DATABASE_DEBUG"); debug == "true" {
			db.DB().LogMode(true)
		}
	}


	log.Info("Connecting with database...")
	err := db.NewConnection(dbUri)
	db.DB().SetLogger(&utils.GormLogger{})
	if err != nil {
		log.Error("Connecting failed: ", err.Error())
	} else {
		log.Info("Connected to the database")
		dbSetup()
	}

	go func() {
		for range dbTicker.Ticker.C {
			reconnected, err := dbTicker.TryReconnect()
			logEntry := utils.AuthLogger().WithField("prefix", "[AUTH DATABASE CONNECT]")
			if err != nil {
				logEntry.Error("Error while trying to connect to database: ", err.Error())
			}
			if reconnected {
				logEntry.Info("Connected to the database")
				dbSetup()
			}
		}
	}()

	authServer := auth.NewAuthorizationServer(db, utils.AuthLogger())
	authServer.Initialize(cfg)

	utils.AuthLogger().Fatal(authServer.Run(cfg))

}