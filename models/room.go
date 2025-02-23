package models

import (
	"encoding/base64"
	"errors"
)

type MsgType string

const (
	Image MsgType = "image"
	Text  MsgType = "text"
)

type MessageRequest struct {
	Type    MsgType `json:"type"`
	Room    string  `json:"room"`
	Nick    string  `json:"nick"`
	Message []byte  `json:"base64Message"`
}

func (r *MessageRequest) Validate() bool {
	return r.Room != "" && r.Nick != "" && len(r.Message) != 0
}

func (m *MessageRequest) ToMessage() (Message, error) {
	switch m.Type {
	case Image:
		data := base64.StdEncoding.EncodeToString(m.Message)
		return Message{Room: m.Room, Nick: m.Nick, Message: []byte(data)}, nil
	case Text:
		return Message{Room: m.Room, Nick: m.Nick, Message: m.Message}, nil
	default:
		return Message{}, errors.New("invalid message type")
	}
}

type Message struct {
	Room    string `json:"room"`
	Nick    string `json:"nick"`
	Message []byte `json:"message"`
}
