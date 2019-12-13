package app

import (
	"FlankiRest/config"
	"FlankiRest/controllers"
	"FlankiRest/database"
	"FlankiRest/logger"
	"FlankiRest/models"
	"FlankiRest/services"
	"FlankiRest/utils"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	stdlogger "log"
	"net/http"
	"os"
	"time"
)

type App struct {
	ApiDB     *database.ApiDatabase
	dbTicker  *utils.DatabaseReconnectTicker
	Router    *mux.Router
	Logger    *logrus.Logger
	AuthCfg   *oauth2.Config // needed for account's controller when setting routing
	AppConfig *config.AppConfig
}

func NewApp(cfg *oauth2.Config) *App {
	return &App{AuthCfg: cfg,ApiDB: &database.ApiDatabase{}}
}

func (app *App) GetDatabaseInstance() *database.ApiDatabase {
	return app.ApiDB
}

type GormLogger struct {}

func (*GormLogger) Print(v ...interface{}) {
	log := logger.GetGlobalLogger()
	if v[0] == "sql" {
		log.WithFields(logrus.Fields{"module": "gorm", "type": "sql", "prefix": "DATABASE"}).Debug(v[3])
	}
	if v[0] == "log" {
		log.WithFields(logrus.Fields{"module": "gorm", "type": "log", "prefix": "DATABASE"}).Debug(v[2])
	}
}

func (app *App) Initialize(apiConfig *config.AppConfig) {

	if app.Logger == nil {
		stdlogger.Fatal("Uninitialized Logger, exiting...")
	}

	app.AppConfig = apiConfig
	if apiConfig == nil {
		app.Logger.Fatal("Application config was not provided")
	}

	log := app.Logger
	log.Info("Initializing app")

	dbUri := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", apiConfig.DBHost, apiConfig.DBUsername, apiConfig.DBName, apiConfig.DBPassword)
	databaseSetup := func() {

		if apiLogger, found := os.LookupEnv("DATABASE_API_LOGGER"); found && apiLogger == "true" {
			app.GetDatabaseInstance().DB().SetLogger(&GormLogger{})
		}
		if dbDebug, found := os.LookupEnv("DATABASE_DEBUG"); found && dbDebug == "true" {
			app.GetDatabaseInstance().DB().LogMode(true)
		}
		app.GetDatabaseInstance().DB().AutoMigrate(&models.Account{}, &models.Lobby{}, &models.Team{}, &models.TeamEntry{}, &models.PlayerStatisticsEntry{}, &services.PasswordReset{})
		app.GetDatabaseInstance().DB().Model(&models.Team{}).AddForeignKey("lobby_id", "lobbies(id)", "CASCADE", "CASCADE")
		app.GetDatabaseInstance().DB().Model(&models.TeamEntry{}).AddForeignKey("team_id", "teams(id)", "CASCADE", "CASCADE")

	}

	log.Info("Connecting with database")
	err := app.GetDatabaseInstance().NewConnection(dbUri)
	if err != nil {
		log.Error("Error while creating connection with Database: " + err.Error())
	} else {
		log.Info("Established connection with database")
		databaseSetup()
	}



	app.dbTicker = utils.NewDatabaseReconnectTicker(app.ApiDB, dbUri, time.Second * 30)
	go func() {
		for range app.dbTicker.Ticker.C {
			reconnected, err := app.dbTicker.TryReconnect()
			logEntry := app.Logger.WithField("prefix", "[DATABSE CONNECT]")
			if err != nil {
				logEntry.Error("Error while trying to connect to database: ", err.Error())
			}
			if reconnected {
				logEntry.Info("Connected to the database")
				databaseSetup()
			}
		}
	}()



	app.SetRouting(config.API_PREFIX)
	if app.Router == nil {
		log.Fatal("app.Router == nil")
	}
}


func (app *App) SetRouting(API_PREFIX string) {
	log := app.Logger
	log.Info("Setting routing for app")
	accountController    := controllers.NewAccountController(app.GetDatabaseInstance(), app.AuthCfg, app.Logger)
	lobbyController      := controllers.NewLobbyController(app.GetDatabaseInstance(), app.Logger)
	playerController     := controllers.NewPlayerController(app.GetDatabaseInstance(), app.Logger)
	statisticsController := controllers.NewStatisticsController(app.GetDatabaseInstance(), app.Logger)
	imageController      := controllers.NewImageController(app.Logger)

	app.Router = mux.NewRouter()
	app.Post(  API_PREFIX + "/user/create",                   accountController.CreateAccount)
	app.Post(  API_PREFIX + "/user/login",                    accountController.LoginAccount)
	app.Patch( API_PREFIX + "/user/me",                       accountController.UpdateAccount)
	app.Delete(API_PREFIX + "/user/me",                       accountController.DeleteAccount)
	app.Get(   API_PREFIX + "/user/me",                       accountController.GetAccount)

	app.Get(   API_PREFIX + "/players",                       playerController.GetAllPlayers)
	app.Get(   API_PREFIX + "/players/{id:[0-9]+}",           playerController.GetPlayerById)
	app.Get(   API_PREFIX + "/players/{id:[0-9]+}/summary",   statisticsController.GetPlayerSummary)
	app.Get(   API_PREFIX + "/players/ranking",               statisticsController.GetPlayersRanking)


	app.Get(   API_PREFIX + "/lobbies/owner",                 lobbyController.OwnerLobby)
	app.Delete(API_PREFIX + "/lobbies/owner",                 lobbyController.DeleteLobby)
	app.Patch( API_PREFIX + "/lobbies/owner",                 lobbyController.UpdateLobby)
	app.Post(  API_PREFIX + "/lobbies/owner/create",          lobbyController.CreateLobby)
	app.Post(  API_PREFIX + "/lobbies/owner/submit",          lobbyController.SubmitResults)
	app.Post(  API_PREFIX + "/lobbies/owner/close",           lobbyController.CloseLobby)
	app.Post(  API_PREFIX + "/lobbies/owner/kick_player",     lobbyController.KickPlayerFromLobby)
	app.Get(   API_PREFIX + "/lobbies/my",                    lobbyController.GetCurrentLobby)

	app.Get(   API_PREFIX + "/lobbies",                       lobbyController.GetAllLobbies)
	app.Get(   API_PREFIX + "/lobbies/results",               lobbyController.Results)
	app.Get(   API_PREFIX + "/lobbies/{id:[0-9]+}",           lobbyController.GetLobbyById)
	app.Post(  API_PREFIX + "/lobbies/{id:[0-9]+}/join",      lobbyController.JoinLobbyTeam)
	app.Post(  API_PREFIX + "/lobbies/my/leave",              lobbyController.LeaveLobby)

	app.Get(   API_PREFIX + "/images/{id:[0-9]+}",			   imageController.GetImageById)
	app.Post(  API_PREFIX + "/images/my",			  		   imageController.UploadImage)
	app.Get(   API_PREFIX + "/images/my",			  		   imageController.GetOwnerImage)

	app.Post(  API_PREFIX + "/remember_password",			   accountController.ResetPasswordRequest)
	app.Post(  API_PREFIX + "/reset_password",			       accountController.ResetPassword)

	var measure bool
	if env ,ok := os.LookupEnv("MEASURE_REQUEST_TIME"); ok && env == "true" {
		measure = true
	} else {
		measure = false
	}
	requestTimer := NewRequestTimer(app.Logger, measure)
	app.Router.Use(mux.CORSMethodMiddleware(app.Router), Oauth2Authentication, requestTimer.RequestTimeMiddleware) //attach JWT auth middleware
}

func (app *App) SetLogger(logger *logrus.Logger) {
	app.Logger = logger
}

func (app *App) SetOauth2Config(cfg *oauth2.Config) {
	app.AuthCfg = cfg
}

func (app *App) Run(enableSSL bool) error {

	defer app.GetDatabaseInstance().Close()

	c := cors.New(cors.Options{
		AllowedMethods: []string{"GET","POST","DELETE","PATCH"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	})
	var err error

	server := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      c.Handler(app.Router),
		Addr: 		   ":" + app.AppConfig.Port,
	}
	if enableSSL {
		app.Logger.Info("Using SSL")
		/*
		m := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("flaneczki.pl"),
			Cache:      autocert.DirCache("/ssl_certs"),
		}
		*/

		if val, ok := os.LookupEnv("SSL_PORT"); ok {
			server.Addr = ":" + val
		} else {
			server.Addr = ":443"
		}


		// saw this here: https://stackoverflow.com/questions/37321760/how-to-set-up-lets-encrypt-for-a-go-server-application
		// will see if it really works, pls do ):
		//go http.ListenAndServe(":http", m.HTTPHandler(nil))

		app.Logger.Info("Application is listening on port " + server.Addr)
		err = server.ListenAndServeTLS("/ssl_certs/cert.pem", "/ssl_certs/privkey.pem")
	}  else {
		app.Logger.Info("Application is listening on port " + server.Addr)
		err = server.ListenAndServe()
	}
	return err
}

func LoggerFuncWrapper(f func(w http.ResponseWriter, r *http.Request))  func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		prefix := fmt.Sprintf("[%s %s]", r.URL.Path, r.Method)
		logger.GetGlobalLogger().WithField("prefix", prefix).Info(r.RemoteAddr)
		f(w,r)
		return
	}
}

// Get wraps the Router for GET method
func (app *App) Get(path string, f func(w http.ResponseWriter, r *http.Request)) {
	app.Router.HandleFunc(path, LoggerFuncWrapper(f)).Methods("GET")
}

// Post wraps the Router for POST method
func (app *App) Post(path string, f func(w http.ResponseWriter, r *http.Request)) {
	app.Router.HandleFunc(path, LoggerFuncWrapper(f)).Methods("POST")
}

// Put wraps the Router for PUT method
func (app *App) Put(path string, f func(w http.ResponseWriter, r *http.Request)) {
	app.Router.HandleFunc(path, LoggerFuncWrapper(f)).Methods("PUT")
}

// Delete wraps the Router for DELETE method
func (app *App) Delete(path string, f func(w http.ResponseWriter, r *http.Request)) {
	app.Router.HandleFunc(path, LoggerFuncWrapper(f)).Methods("DELETE")
}

// Patch  wraps the Router for DELETE method
func (app *App) Patch(path string, f func(w http.ResponseWriter, r *http.Request)) {
	app.Router.HandleFunc(path, LoggerFuncWrapper(f)).Methods("PATCH")
}





