package models

import (
	"encoding/base64"
	"errors"
)

type MsgType string

const (
	File MsgType = "file"
	Text MsgType = "text"
)

type MessageRequest struct {
	Type     MsgType `json:"type"`
	Room     string  `json:"room"`
	Nick     string  `json:"nick"`
	FileName string  `json:"filename"`
	Message  []byte  `json:"base64message"`
}

func (r *MessageRequest) Validate() bool {
	return r.Room != "" && r.Nick != "" && len(r.Message) != 0
}

func (m *MessageRequest) ToMessage() (Message, error) {
	switch m.Type {
	case File:
		data := base64.StdEncoding.EncodeToString(m.Message)
		return Message{Room: m.Room, Nick: m.Nick, Message: []byte(data), FileName: m.FileName}, nil
	case Text:
		return Message{Room: m.Room, Nick: m.Nick, Message: m.Message}, nil
	default:
		return Message{}, errors.New("invalid message type")
	}
}

type Message struct {
	Room     string `json:"room"`
	Nick     string `json:"nick"`
	Message  []byte `json:"message"`
	FileName string `json:"filename"`
}
