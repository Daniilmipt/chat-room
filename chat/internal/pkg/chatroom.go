package pkg

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
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

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	self peer.ID
	Room string
	Nick string

	writer *bufio.Writer
}

// ChatMessage gets converted to/from JSON and sent in the body of pubsub messages.
type ChatMessage struct {
	Message    string
	SenderID   string
	SenderNick string
}

// JoinChatRoom tries to subscribe to the PubSub topic for the room name, returning
// a ChatRoom on success.
func JoinChatRoom(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, nick string, room string, writer *bufio.Writer, errCh chan<- error) *ChatRoom {
	topic, err := ps.Join(topicName(room))
	if err != nil {
		return nil
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return nil
	}

	cr := &ChatRoom{
		ctx:      ctx,
		ps:       ps,
		topic:    topic,
		sub:      sub,
		self:     selfID,
		Nick:     nick,
		Room:     room,
		Messages: make(chan *ChatMessage, ChatRoomBufSize),
		writer:   writer,
	}

	go cr.readLoop(errCh)
	return cr
}

// Publish sends a message to the pubsub topic.
func (cr *ChatRoom) Publish(message string) error {
	m := ChatMessage{
		Message:    message,
		SenderID:   cr.self.String(),
		SenderNick: cr.Nick,
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return cr.topic.Publish(cr.ctx, msgBytes)
}

func (cr *ChatRoom) readLoop(errCh chan<- error) {
	defer close(cr.Messages)
	for {
		msg, err := cr.sub.Next(cr.ctx)
		if err != nil {
			errCh <- err
			return
		}

		cm := &ChatMessage{}
		if err = json.Unmarshal(msg.Data, cm); err != nil {
			continue
		}

		if err = cr.writeInFile(cm); err != nil {
			errCh <- err
			return
		}

		if msg.ReceivedFrom == cr.self {
			continue
		}

		cr.Messages <- cm
	}
}

func (cr *ChatRoom) writeInFile(cm *ChatMessage) error {
	logEntry := fmt.Sprintf("%s: %s: %s\n", cr.Nick, cm.Message, time.Now())
	if _, err := cr.writer.WriteString(logEntry); err != nil {
		return err
	}

	if err := cr.writer.Flush(); err != nil {
		return err
	}

	return nil
}

func topicName(room string) string {
	return "chat-room:" + room
}

func SendMessage(cr *ChatRoom, lgr *zap.Logger) {
	logger := lgr.With(zap.String("nick", cr.Nick), zap.String("room", cr.Room))

	scanner := bufio.NewScanner(os.Stdin)
	for {
		if scanner.Scan() {
			message := scanner.Text()
			if message == "" {
				continue
			}

			logger.Info("received message", zap.String("message", message))

			err := cr.Publish(message)
			if err != nil {
				logger.Error("failed to send message",
					zap.String("message", message),
					zap.Error(err),
				)
				continue
			}

			logger.Info("message was published",
				zap.String("message", message),
			)
		} else {
			if err := scanner.Err(); err != nil {
				logger.Error("failed to get scanner error", zap.Error(err))
			}
			break
		}
	}
}
