package websocketchat

import (
	"context"
	"sync"
)

type ChatRoom struct {

	// owner of the room, allows him to kick people out, mute them and close the room (not all of this is implemented right now)
	OwnerID uint

	// clients map(sneaky set)
	Clients map[*Client] bool

	// queue for outgoing  messages to clients
	OutMessages chan *Message

	// queue for Clients that want to join chat room
	Register chan *Client

	// queue for Clients leaving chat room
	Unregister chan *Client

	// chanel for closing chanel launched by Run() method
	Close chan struct{}

	// mutex for securing clients' map
	clientMutex sync.RWMutex

}

func NewChatRoom() *ChatRoom {
	room := &ChatRoom{}
	room.Clients = map[*Client] bool {}
	room.OutMessages = make(chan *Message, 100)
	room.Register = make(chan *Client, 100)
	room.Unregister = make(chan *Client, 100)
	room.Close = make(chan struct{})
	return room
}

func (room *ChatRoom) Users() []string {
	// read lock
	room.clientMutex.RLock()
	defer room.clientMutex.RUnlock()

	count := len(room.Clients)
	users := make([]string, 0, count)
	for k,v := range room.Clients {
		if v {
			users = append(users, k.User.Nickname)
		}
	}
	return users
}

func (room *ChatRoom) DispatchMessage(ctx context.Context, msg *Message) {

	switch msg.Action {
	case "message":
		room.OutMessages <- msg
	case "users":
		client := ctx.Value("client").(*Client)
		msg.Text = ""
		msg.Data = room.Users()
		client.Send <- msg
	case "kick":
		client := ctx.Value("client").(*Client)
		if !room.IsUserPrivileged(client.User) {
			break
		}
		// message should contain nickname of the user to be kicked out
		user := msg.Text
		for k,_ := range room.Clients {
			if k.User.Nickname == user {
				msg.Text = "You have been kicked out of the room!"
				k.Send <- msg
				room.Unregister <- k
				break
			}
		}
	}
}

func (room *ChatRoom) IsUserPrivileged(user User) bool {
	if user.ID == room.OwnerID {
		return true
	}
	return false
}

// Running room actions, should be started as goroutine
func (room *ChatRoom) Run() {
	for {
		select {
		case msg := <- room.OutMessages:
			for c, _ := range room.Clients {
				c.Send <- msg
			}

		case c := <- room.Register:
			c.Room = room
			room.Clients[c] = true

		case c := <- room.Unregister:
			// write lock
			room.clientMutex.Lock()
			if _, ok := room.Clients[c]; ok {
				delete(room.Clients, c)
				close(c.Send)
			}
			room.clientMutex.Unlock()

		case <- room.Close:
			for c,_ := range room.Clients {
				close(c.Send)
			}
			// let the grabage collector clean clients' set
			room.Clients = map[*Client] bool{}
			return
		}
	}
}
