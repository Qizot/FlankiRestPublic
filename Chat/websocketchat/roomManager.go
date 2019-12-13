package websocketchat

import (
	"errors"
	"sync"
)

var (
	RoomNameTaken     = errors.New("room name is already taken")
	RoomNotFound      = errors.New("room with given name has not been found")
	RoomLimitReached  = errors.New("can't create any new rooms, limit reached")
)

var (
	MaxLobbiesLimit = 100
)

type RoomManager struct {
	rooms map[string] *ChatRoom
	mutex sync.Mutex
}

func NewRoomManager() *RoomManager {
	return &RoomManager{map[string] *ChatRoom{}, sync.Mutex{}}
}

func (manager *RoomManager) CreateNewRoom(name string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if len(manager.rooms) >= MaxLobbiesLimit {
		return RoomLimitReached
	}
	if _, found := manager.rooms[name]; found {
		return RoomNameTaken
	}
	room := NewChatRoom()
	manager.rooms[name] = room
	go room.Run()
	return nil
}

func (manager *RoomManager) CloseRoom(name string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if room, found := manager.rooms[name]; !found {
		return RoomNotFound
	} else {
		room.Close <- struct{}{}
		delete(manager.rooms, name)
		return nil
	}
}

func (manager *RoomManager) SelfCloseRoom(room *ChatRoom) (err error) {
	for k,v := range manager.rooms {
		if v == room {
			err = manager.CloseRoom(k)
			return
		}
	}
	return
}

func (manager *RoomManager) GetRoom(name string) *ChatRoom {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	return manager.rooms[name]
}

func (manager *RoomManager) ListRooms() []string {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	rooms := make([]string, 0, len(manager.rooms))
	for k,_ := range manager.rooms {
		rooms = append(rooms, k)
	}
	return rooms
}


