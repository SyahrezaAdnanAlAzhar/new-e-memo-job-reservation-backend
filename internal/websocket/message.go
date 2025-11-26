package websocket

import (
	"encoding/json"
)

type Message struct {
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
}

func NewMessage(event string, payload interface{}) ([]byte, error) {
	msg := Message{
		Event:   event,
		Payload: payload,
	}
	return json.Marshal(msg)
}
