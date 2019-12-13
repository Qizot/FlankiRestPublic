package controllers

import (
	"FlankiRest/database"
	"FlankiRest/errors"
	"FlankiRest/models"
	"FlankiRest/services"
	u "FlankiRest/utils"
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"net/http"
	"time"
)

type AccountController struct {
	DB *database.ApiDatabase
	Cfg *oauth2.Config
	logger *logrus.Logger
}

func NewAccountController(db *database.ApiDatabase, cfg *oauth2.Config, logger *logrus.Logger) *AccountController {
	return &AccountController{db, cfg, logger}
}

func (controller *AccountController) Logger() *logrus.Logger {
	return controller.logger
}

type Credentials struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func (controller *AccountController) CreateAccount(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	account := &models.Account{}
	err := json.NewDecoder(r.Body).Decode(&account)

	// comparing unmarshaled struct to empty struct in case when json was valid but didn't fill the struct
	if err != nil || *account == (models.Account{}) {
		u.ApiErrorResponse(w, errors.BadJsonRequestFormat)
		return
	}
	err = account.Create(db)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	u.SimpleRespond(w, u.TextMessage("Account has been created, you can now log in into your account"))
	return
}

func (controller *AccountController) LoginAccount(w http.ResponseWriter, r *http.Request) {

	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil || *creds == (Credentials{}) {
		u.ApiErrorResponse(w, errors.BadJsonRequestFormat)
		return
	}

	token, err := controller.Cfg.PasswordCredentialsToken(context.Background(), creds.Email, creds.Password)
	if err != nil {
		if serr, ok := err.(*oauth2.RetrieveError); ok {
			if serr.Response.StatusCode == 401 {
				u.ApiErrorResponse(w, errors.UnauthorizedAccount)
				return
			}
		}
		controller.Logger().WithField("prefix", "[LOGIN]").Error(err.Error())
		u.ApiErrorResponse(w, errors.New("Authentication server internal error",500))
		return
	}

	token.RefreshToken = "" // no refresh tokens because I'm to lazy to implement another REST path
	data := map[string]interface{} {
		"access_token":  token.AccessToken,
		"token_type":    token.TokenType,
		"expires_in": int64(token.Expiry.Sub(time.Now()).Seconds()),
	}
	u.SimpleRespond(w, data)
	return
}

func (controller *AccountController) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	requestedAccountChange := &models.UpdateAccount{}
	err := json.NewDecoder(r.Body).Decode(requestedAccountChange)

	// if request is default to empty UpdateAccount struct it means that request is invalid
	if err != nil || requestedAccountChange.IsEmpty() {
		u.ApiErrorResponse(w, errors.BadJsonRequestFormat)
		return
	}
	id, err := u.GetUserIdFromContext(r.Context())
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	requestedAccountChange.ID = id

	err = requestedAccountChange.Update(db)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	u.SimpleRespond(w, u.TextMessage("Account has been updated"))
	return
}

func (controller *AccountController) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()

	id, err := u.GetUserIdFromContext(r.Context())
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}

	account, err := models.GetAccountById(db, id)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	// user should not be able to delete account while he is playing
	// this prevents from further bugs like submitting results of deleted account ect.
	if account.Playing {
		u.ApiErrorResponse(w, errors.New("User is playing, can't delete this account", 400))
		return
	}
	err = db.Delete(account).Error
	if err != nil {
		u.ApiErrorResponse(w, errors.DatabaseError(err))
		return
	}
	u.SimpleRespond(w, u.TextMessage("Account has been deleted"))
	return
}

func (controller *AccountController) GetAccount(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()

	id, err := u.GetUserIdFromContext(r.Context())
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}

	account, err := models.GetAccountById(db, id)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	account.Password = ""
	userStatistics, err := models.GetPlayersSummary(db, id)
	var summary *models.QuickSummary
	if userStatistics != nil {
		summary = userStatistics.GetQuickSummary()
	} else {
		summary = nil
	}
	response := map[string] interface{} {"summary": summary, "account": account}
	u.SimpleRespond(w, response)
	return
}

func (controller *AccountController) ResetPasswordRequest(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	type Email struct {
		Email string `json:"email"`
	}
	email := &Email{}
	err := json.NewDecoder(r.Body).Decode(email)
	if err != nil {
		u.ApiErrorResponse(w, errors.BadJsonRequestFormat)
		return
	}

	err = services.EmailPasswordResetRequest(db, email.Email)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}

	u.SimpleRespond(w, u.TextMessage("Further instructions should have been sent to your email"))
	return
}

func (controller *AccountController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	db := controller.DB.DB()
	request := services.ResetRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		u.ApiErrorResponse(w, errors.BadJsonRequestFormat)
		return
	}

	err = services.ResetPassword(db, request)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}

	u.SimpleRespond(w, u.TextMessage("Password has been changed!"))
	return
}
