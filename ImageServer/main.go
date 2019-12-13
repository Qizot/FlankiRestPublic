package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/x-cray/logrus-prefixed-formatter"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	HomeFolder  = "."
	logger *logrus.Logger
	logEntry *logrus.Entry
)

func init() {
	logger = logrus.New()
	customFormatter := new(prefixed.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.ForceColors = true
	customFormatter.ForceFormatting  = true
	customFormatter.FullTimestamp = true
	logger.SetFormatter(customFormatter)
	logger.Level = logrus.DebugLevel
	logEntry = logger.WithField("prefix","[IMAGE SERVER]")

}

func RespondWithStatus(w http.ResponseWriter, code int, data map[string] string) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		logEntry.Error("Got error while responding: " + err.Error())
	}
}

func Message(msg string) map[string] string {
	return map[string] string {"message": msg}
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	const MAX_SIZE = 3 * (1 << (10 * 2))
	r.Body = http.MaxBytesReader(w,r.Body, MAX_SIZE)
	defer r.Body.Close()

	if contentType := r.Header.Get("Content-Type"); contentType != "image/jpeg" && contentType != "image/png" {
		logEntry.Error("Invalid content type: ", contentType)
		RespondWithStatus(w, 400, Message("Invalid content type: " + contentType))
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logEntry.Error("Error while reading request: " + err.Error())
		RespondWithStatus(w, 400, Message(err.Error()))
		return
	}
	vars := mux.Vars(r)
	f, err := os.Create("./images/" + vars["id"])
	defer f.Close()
	f.Write(b)
	RespondWithStatus(w,200, Message("File has been uploaded!"))
	logEntry.Debug("Uploaded file with id: ", vars["id"])
	return
}

func main() {

	router := mux.NewRouter()
	router.PathPrefix("/images/").Handler(
		http.StripPrefix("/images/", http.FileServer(http.Dir("./images/")))).Methods("GET")
	router.HandleFunc("/upload/{id:[0-9]+}", uploadFile).Methods("POST")
	logger.Info("Image server running on port: 5555")
	logEntry.Fatal(http.ListenAndServe(":5555", router))
}