package websocketchat

import (
	"encoding/json"
	"net/http"
)

func ChatErrorResponse(w http.ResponseWriter, e string, status int) {

	Logger().Error(e)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	data := map[string] string {"message": e}
	_ = json.NewEncoder(w).Encode(data)
}

func SimpleRespond(w http.ResponseWriter, text interface{}) {
	w.Header().Add("Content-Type", "application/json")
	data := map[string]interface{} {"message": text}
	_ = json.NewEncoder(w).Encode(data)
}
