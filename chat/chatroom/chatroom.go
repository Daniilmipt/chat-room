package chatroom

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

// ChatRoom represents a subscription to a single PubSub topic. Messages
// can be published to the topic with ChatRoom.Publish, and received
// messages are pushed to the Messages channel.
type ChatRoom struct {
	ps    *pubsub.PubSub
	Topic *pubsub.Topic
	Sub   *pubsub.Subscription

	Self peer.ID
	Room string
	Nick string

	writer *bufio.Writer
	file   *os.File
}

type ChatMessage struct {
	Message    []byte
	SenderNick string
	FileName   string
}

type ChatRoomFileOptions struct {
	Writer *bufio.Writer
	File   *os.File
}

func JoinChatRoom(ctx context.Context, logger *zap.Logger, ps *pubsub.PubSub, selfID peer.ID, room, nick string, fileOpts ChatRoomFileOptions) (*ChatRoom, error) {
	topic, err := ps.Join(topicName(room))
	if err != nil {
		return nil, err
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	cr := &ChatRoom{
		ps:     ps,
		Topic:  topic,
		Sub:    sub,
		Self:   selfID,
		Room:   room,
		writer: fileOpts.Writer,
		file:   fileOpts.File,
		Nick:   nick,
	}
	return cr, nil
}

func (cr *ChatRoom) Publish(ctx context.Context, message string) error {
	m := ChatMessage{
		Message:    []byte(message),
		SenderNick: cr.Nick,
	}

	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return cr.Topic.Publish(ctx, msgBytes)
}

func (cr *ChatRoom) PublishWithFile(ctx context.Context, filename string, message []byte) error {
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

func (cr *ChatRoom) ReadLoop(ctx context.Context, logger *zap.Logger) {
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

		if msg.ReceivedFrom == cr.Self {
			continue
		}
	}
}

func (cr *ChatRoom) writeInFile(cm *ChatMessage) error {
	fileMsg := fmt.Sprintf("%s: %s: %s: %s\n", cm.SenderNick, cm.FileName, cm.Message, time.Now())
	if _, err := cr.writer.WriteString(fileMsg); err != nil {
		return err
	}

	if err := cr.writer.Flush(); err != nil {
		return err
	}

	return nil
}

func (cr *ChatRoom) SendMessageWithFile(ctx context.Context, logger *zap.Logger, nick, filename string, message []byte) {
	logger.Info("received message", zap.ByteString("message", message))
	defer logger.Info("message was published", zap.ByteString("message", message))

	if err := cr.PublishWithFile(ctx, filename, message); err != nil {
		logger.Error("failed to send message",
			zap.ByteString("message", message),
			zap.Error(err),
		)
		return
	}
}

func (cr *ChatRoom) SendMessage(ctx context.Context, logger *zap.Logger, nick, message string) {
	logger.Info("received message", zap.String("message", message))
	defer logger.Info("message was published", zap.String("message", message))

	if err := cr.Publish(ctx, message); err != nil {
		logger.Error("failed to send message",
			zap.String("message", message),
			zap.Error(err),
		)
		return
	}
}

func (cr *ChatRoom) Close() error {
	cr.Sub.Cancel()

	if err := cr.file.Close(); err != nil {
		return err
	}
	return cr.Topic.Close()
}

func topicName(room string) string {
	return "chat-room:" + room
}
