package authorization

import (
	"AuthorizationServer/database"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
	"net/http"
	"strconv"
	"time"
)


type AuthorizationServer struct {
	Manager     *manage.Manager
	ClientStore *store.ClientStore
	TokenServer *server.Server
	Router      *mux.Router
	DB          *database.AuthDatabase // db connection is needed for when it is needed to check user's existence in db and validate his credentials
	Logger	    *logrus.Logger
}

func NewAuthorizationServer(db *database.AuthDatabase, logger *logrus.Logger) *AuthorizationServer {
	return &AuthorizationServer{DB: db, Logger: logger}
}

var CustomAuthorizationCodeTokenCfg = &manage.Config{AccessTokenExp: time.Hour * 24 * 7, RefreshTokenExp: time.Second * 20, IsGenerateRefresh: false}

func (authserver *AuthorizationServer) Initialize(cfg *AuthServerConfig) {
	log := authserver.Logger
	manager := manage.NewDefaultManager()
	//manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	manager.SetPasswordTokenCfg(CustomAuthorizationCodeTokenCfg)

	// token store
	manager.MustTokenStorage(NewStoreWithDB(authserver.DB, 3600, 30, authserver.Logger))

	clientStore := store.NewClientStore()

	for _, client := range cfg.Clients {
		err := clientStore.Set(client.ID, &client)
		if err != nil {
			log.Fatal("Authorization server failed to load clients")
		}
	}
	manager.MapClientStorage(clientStore)

	srv := server.NewServer(server.NewConfig(), manager)
	srv.SetPasswordAuthorizationHandler(PasswordAuthenticationHandler(authserver.DB))

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		re = &errors.Response{}
		re.Error = err
		if err == ErrUnauthorizedAccount {
			re.StatusCode = http.StatusUnauthorized
		} else {
			re.StatusCode = 500
		}
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {

	})

	router := mux.NewRouter()

	router.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		err := srv.HandleTokenRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("POST")

	router.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		token, err := srv.ValidationBearerToken(r)
		if err != nil {
			if err == ErrDatabaseError {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		userID, _ := strconv.Atoi(token.GetUserID())
		data := map[string]interface{}{
			"expires_in": int64(token.GetAccessCreateAt().Add(token.GetAccessExpiresIn()).Sub(time.Now()).Seconds()),
			"client_id":  token.GetClientID(),
			"user_id":   userID,
		}
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		err = e.Encode(data)
		authserver.Logger.WithField("prefix", "[AUTH SERVER]").Debug("Authorizing user with id: ", userID)
	}).Methods("POST")

	authserver.Manager = manager
	authserver.ClientStore = clientStore
	authserver.TokenServer = srv
	authserver.Router = router
}

func (authserver *AuthorizationServer) Run(cfg *AuthServerConfig) error {
	authserver.Logger.Info("Authorization server is listening on port " + cfg.Port)
	err := http.ListenAndServe(":" + cfg.Port, authserver.Router)
	return err
}

func PasswordAuthenticationHandler(db *database.AuthDatabase) func(string, string) (string, error) {
	return func(username, password string) (userID string, err error) {
		if db == nil {
			err = fmt.Errorf("Database not found while trying to authorize user")
			return
		}
		id, err := FetchUserID(db.DB(), username, password)
		if err != nil {
			return
		}
		userID = fmt.Sprint(id)
		return
	}
}

