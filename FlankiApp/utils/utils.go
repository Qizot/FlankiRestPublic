package utils

import (
	"FlankiRest/database"
	"FlankiRest/errors"
	"FlankiRest/logger"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)


func GetUserIdFromContext(ctx context.Context) (uint,error) {
	if id, ok := ctx.Value("user").(uint); !ok {
		return 0, errors.New("Couldn't retrieve player's id from context", 500)
	} else {
		return id, nil
	}
}

type ErrMessage struct {
	Message string `json:"message"`
}

func DecodeErrMessage(b []byte) ErrMessage {
	msg := ErrMessage{}
	_ = json.NewDecoder(bytes.NewBuffer(b)).Decode(&msg)
	return msg
}

func TextMessage(message interface{}) map[string] interface{} {
	return map[string] interface{} {"message": message}
}

func ApiErrorResponse(w http.ResponseWriter, e error) {

	w.Header().Add("Content-Type", "application/json")


	apiErr, ok := e.(*errors.ApiError)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(TextMessage("Uknown error: " + e.Error()))
		return
	}
	logger.GetGlobalLogger().WithField("prefix", "[RESPONSE]").Error(apiErr.Error())
	w.WriteHeader(apiErr.HttpCode)

	data := map[string] string {"message": apiErr.Message}
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println("Got error while responding: " + err.Error())
	}
}

func SimpleRespond(w http.ResponseWriter, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println("Got error while responding: " + err.Error())
	}
}

type DatabaseReconnectTicker struct {
	db *database.ApiDatabase
	uri string
	Ticker *time.Ticker
}

func NewDatabaseReconnectTicker(conn *database.ApiDatabase,uri string,  interval time.Duration) *DatabaseReconnectTicker {
	drt := &DatabaseReconnectTicker{conn,uri, time.NewTicker(interval)}
	return drt
}

func (drt *DatabaseReconnectTicker) TryReconnect() (bool, error) {
	if err := drt.db.DB().DB().Ping(); err != nil {
		_ = drt.db.Close()
		err = drt.db.NewConnection(drt.uri)
		if err != nil {
			return false, err
		} else {
			return true, err
		}
	}
	return false, nil
}




