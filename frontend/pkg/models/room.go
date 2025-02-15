package models

import (
	"encoding/base64"

	"github.com/pkg/errors"
)

type MessageRequest struct {
	Room    string `json:"room"`
	Nick    string `json:"nick"`
	Message string `json:"base64Message"`
}

func (r *MessageRequest) Validate() bool {
	return r.Room != "" && r.Nick != "" && len(r.Message) != 0
}

type Message struct {
	Room    string `json:"room"`
	Nick    string `json:"nick"`
	Message []byte `json:"message"`
}

func (m *MessageRequest) ToMessage() (Message, error) {
	data, err := base64.StdEncoding.DecodeString(m.Message)
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to convert string base64 message to byte array")
	}

	return Message{Room: m.Room, Nick: m.Nick, Message: data}, nil
}
