package websocketchat

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type ChatServer struct {
	roomManager *RoomManager
	router *mux.Router
}


func LoggerFuncWrapper(f func(w http.ResponseWriter, r *http.Request))  func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		prefix := fmt.Sprintf("[%s %s]", r.URL.Path, r.Method)
		RawLogger().WithField("prefix", prefix).Info(r.RemoteAddr)
		f(w,r)
		return
	}
}

// creates default ChatServer containing one room: 'general' where all users without specified room will be redirected
func NewChatServer() *ChatServer {
	server := ChatServer{}
	server.roomManager = NewRoomManager()
	connectionController := NewConnectController(server.roomManager)
	server.router = mux.NewRouter()

	server.Get("/chat/join/{name}", connectionController.JoinRoom)
	server.Get("/chat/rooms", connectionController.ListRooms)
	server.Post("/chat/create/{name}", connectionController.CreateRoom)
	server.Post("/chat/close/{name}", connectionController.CloseRoom)

	_ = server.roomManager.CreateNewRoom("general")
	return &server
}

func (server *ChatServer) Run(address string, enableSLL string) error {
	// SLL not needed yet

	RawLogger().Info("Chat running on ", address)
	return http.ListenAndServe(address, server.router)
}

func (server *ChatServer) Get(path string, f func(w http.ResponseWriter, r *http.Request)) {
	server.router.HandleFunc(path, LoggerFuncWrapper(f)).Methods("GET")
}

// Post wraps the Router for POST method
func (server *ChatServer) Post(path string, f func(w http.ResponseWriter, r *http.Request)) {
	server.router.HandleFunc(path, LoggerFuncWrapper(f)).Methods("POST")
}

