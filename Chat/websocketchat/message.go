package websocketchat

import "time"


var (
	Actions = []string{"message", "users"}
)

type Message struct {
	Nickname string      `json:"nickname"`
	Action   string      `json:"action,omitempty"`
	Text     string      `json:"text"`
	Time     time.Time   `json:"time"`
	Data     interface{} `json:"data"`
}

func (msg *Message) HasValidAction() bool {
	for _,a := range Actions {
		if a == msg.Action {
			return true
		}
	}
	return false
}