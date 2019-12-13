package websocketchat

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

type Client struct {
	socket *websocket.Conn
	Room *ChatRoom
	Send chan *Message
	User User
}

func (client *Client) ReadRoutine() {
	defer client.socket.Close()

	err := client.socket.ReadJSON(&client.User)
	if err != nil {
		fmt.Println("Couldn't read client's user information: ", err.Error())
		return
	}

	err = GetFlankiChecker.Authorize(&client.User)
	if err != nil {
		fmt.Println("Invalid user: ", err.Error())
		return
	}

	err = GetFlankiChecker.FetchUserInformation(&client.User)
	if err != nil {
		fmt.Println("Error while fetching nickname", err.Error())
		return
	}
	client.User.Connected = time.Now()

	for {
		var message *Message
		err := client.socket.ReadJSON(&message)
		if err != nil {
			return
		}

		// overwrite message's nickname with client's user nickname
		message.Nickname = client.User.Nickname
		message.Time = time.Now()
		ctx := context.Background()
		ctx = context.WithValue(ctx, "client", client)
		client.Room.DispatchMessage(ctx, message)
	}
}

func (client *Client) WriteRoutine() {
	defer client.socket.Close()
	for msg := range client.Send {
		err := client.socket.WriteJSON(msg)
		if err != nil {
			return
		}
	}
}

