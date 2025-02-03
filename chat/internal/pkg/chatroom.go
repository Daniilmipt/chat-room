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

const (
	chatRoomBufSize = 128

	topicNameHeader = "chat-room"
)

type ChatRoom struct {
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

type ChatMessage struct {
	Message    string
	SenderID   string
	SenderNick string
}

func JoinChatRoom(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, nick string, room string, writer *bufio.Writer, errCh chan<- error) *ChatRoom {
	topic, err := ps.Join("chat-room:" + room)
	if err != nil {
		return nil
	}

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
		Messages: make(chan *ChatMessage, chatRoomBufSize),
		writer:   writer,
	}

	go cr.readLoop(errCh)
	return cr
}

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
	return topicNameHeader + ":" + room
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
