package websocketchat

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
)

const (
	socketBufferSize = 1024
	messageBufferSize = 256
)

type ConnectController struct {
	roomManager *RoomManager
	upgrader  *websocket.Upgrader
}


func NewConnectController(manager *RoomManager) *ConnectController{
	upgrader := &websocket.Upgrader{ReadBufferSize: socketBufferSize,
		WriteBufferSize: socketBufferSize}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true;
	}
	return &ConnectController{manager, upgrader}
}

func (controller *ConnectController) JoinRoom(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	room := controller.roomManager.GetRoom(name)
	if room == nil {
		ChatErrorResponse(w, "room not found", 404)
		return
	}

	socket, err := controller.upgrader.Upgrade(w,r,nil)
	if err != nil {
		Logger().Error("socker error: ", err.Error())
		return
	}

	client := &Client{socket, room, make(chan *Message, 10), User{}}
	//TODO we should wait for client to register maybe? Or we don't really have to?
	room.Register <- client
	defer func() { room.Unregister <- client }()

	go client.WriteRoutine()
	client.ReadRoutine()
}

func (controller *ConnectController) CreateRoom(w http.ResponseWriter, r *http.Request) {

	user := User{Token: r.Header.Get("Authorization")}
	if err := GetFlankiChecker.Authorize(&user); err != nil {
		ChatErrorResponse(w, "unauthorized user", 401)
		return
	}
	name := mux.Vars(r)["name"]
	room := controller.roomManager.GetRoom(name)
	if room != nil {
		ChatErrorResponse(w, "room's name is already taken", 400)
		return
	}

	err := controller.roomManager.CreateNewRoom(name)
	if err != nil {
		ChatErrorResponse(w, err.Error(), 400)
		return
	}
	// assign the owner to newly created room
	controller.roomManager.GetRoom(name).OwnerID = user.ID
	SimpleRespond(w, "created new chatroom")
}

func (controller *ConnectController) CloseRoom(w http.ResponseWriter, r *http.Request) {
	user := User{Token: r.Header.Get("Authorization")}
	if err := GetFlankiChecker.Authorize(&user); err != nil {
		ChatErrorResponse(w, "unauthorized user", 401)
		return
	}
	name := mux.Vars(r)["name"]
	room := controller.roomManager.GetRoom(name)
	if room == nil {
		ChatErrorResponse(w, "room not found", 404)
		return
	}
	if room.OwnerID != user.ID {
		ChatErrorResponse(w, "user was not an owner of given room", 401)
		return
	}
	err := controller.roomManager.CloseRoom(name)
	if err != nil {
		ChatErrorResponse(w, "whoopsie: " + err.Error(), 500)
		return
	}
	SimpleRespond(w, "room has been closed")
	return
}

func (controller *ConnectController) ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms := controller.roomManager.ListRooms()
	err := json.NewEncoder(w).Encode(rooms)
	if err != nil {
		Logger().Error(err.Error())
	}
}


