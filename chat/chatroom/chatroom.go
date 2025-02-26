package pkg

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// ChatRoomBufSize is the number of incoming messages to buffer for each topic.
const ChatRoomBufSize = 128

// ChatRoom represents a subscription to a single PubSub topic. Messages
// can be published to the topic with ChatRoom.Publish, and received
// messages are pushed to the Messages channel.
type ChatRoom struct {
	// Messages is a channel of messages received from other peers in the chat room
	Messages chan *ChatMessage

	ps    *pubsub.PubSub
	Topic *pubsub.Topic
	Sub   *pubsub.Subscription

	self peer.ID
	Room string
	Nick string

	writer *bufio.Writer
}

type ChatMessage struct {
	Message    []byte
	SenderNick string
	FileName   string
}

func JoinChatRoom(ctx context.Context, logger *zap.Logger, ps *pubsub.PubSub, selfID peer.ID, room, nick string, writer *bufio.Writer) (*ChatRoom, error) {
	topic, err := ps.Join(topicName(room))
	if err != nil {
		return nil, err
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	cr := &ChatRoom{
		ps:       ps,
		Topic:    topic,
		Sub:      sub,
		self:     selfID,
		Room:     room,
		Messages: make(chan *ChatMessage, ChatRoomBufSize),
		writer:   writer,
		Nick:     nick,
	}

	go cr.readLoop(ctx, logger)
	return cr, nil
}

func (cr *ChatRoom) Publish(ctx context.Context, filename string, message []byte) error {
	m := ChatMessage{
		Message:    message,
		SenderNick: cr.Nick,
		FileName:   filename,
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return cr.Topic.Publish(ctx, msgBytes)
}

func (cr *ChatRoom) readLoop(ctx context.Context, logger *zap.Logger) {
	defer close(cr.Messages)

	for {
		msg, err := cr.Sub.Next(ctx)
		if err != nil { // TODO error "subrscription canceled"
			logger.Error("failed to read next message in room", zap.Error(err))
			return
		}

		if msg == nil {
			logger.Error("get nil message from topic", zap.Any("chat_room", cr))
			return
		}

		cm := &ChatMessage{}
		if err = json.Unmarshal(msg.Data, cm); err != nil {
			logger.Error("failed to unmarshal message in room", zap.Error(err), zap.ByteString("message", msg.Data))
			continue
		}

		if err = cr.writeInFile(cm); err != nil {
			logger.Error("failed to write message in file in room", zap.Error(err), zap.Any("message", cm))
			continue
		}

		if msg.ReceivedFrom == cr.self {
			continue
		}

		cr.Messages <- cm
	}
}

func (cr *ChatRoom) writeInFile(cm *ChatMessage) error {
	logEntry := fmt.Sprintf("%s: %s: %s: %s\n", cr.Nick, cm.FileName, cm.Message, time.Now())
	if _, err := cr.writer.WriteString(logEntry); err != nil {
		return err
	}

	if err := cr.writer.Flush(); err != nil {
		return err
	}

	return nil
}

func (cr *ChatRoom) SendMessage(ctx context.Context, logger *zap.Logger, nick, filename string, message []byte) {
	logger.Info("received message", zap.ByteString("message", message))
	defer logger.Info("message was published", zap.ByteString("message", message))

	if err := cr.Publish(ctx, filename, message); err != nil {
		logger.Error("failed to send message",
			zap.ByteString("message", message),
			zap.Error(err),
		)
		return
	}

}

func (cr *ChatRoom) Close() error {
	cr.Sub.Cancel()
	return cr.Topic.Close()
}

func topicName(room string) string {
	return "chat-room:" + room
}
