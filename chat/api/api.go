package api

import (
	"bufio"
	chroom "chatroom/chat/chatroom"
	"chatroom/chat/config"
	"chatroom/chat/service"
	"chatroom/pkg"
	"context"
	"fmt"
	"os"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type User struct {
	Nick string
	Room string
}

type Handler struct {
	cr map[User]*chroom.ChatRoom

	pubsub *pubsub.PubSub
	host   host.Host

	logger *zap.Logger

	msgFile *os.File
	writer  *bufio.Writer
}

func NewHandler(ctx context.Context, logger *zap.Logger, cfg config.Config) Handler {
	// create messages dir storage if not exists
	if err := os.Mkdir(pkg.MsgDir, os.ModePerm); err != nil {
		logger.Info("can not create messages dir", zap.Error(err))
	}

	s := service.NewService(logger, config.Config{Host: cfg.Host, Port: cfg.Port})

	pubsub, host, err := s.NewPubSub(ctx)
	if err != nil {
		logger.Error("failed to get pubsub", zap.Error(err), zap.Any("p2p-host", host))
	}

	return Handler{
		pubsub: pubsub,
		host:   host,
		logger: logger,
		cr:     make(map[User]*chroom.ChatRoom),
	}
}

func (h *Handler) JoinRoom(ctx context.Context, room, nick string) (*chroom.ChatRoom, error) {
	usr := User{Room: room, Nick: nick}
	if cr, ok := h.cr[usr]; ok {
		return cr, nil
	}

	f, w, err := messageLogWritter(room)
	if err != nil {
		h.logger.Error("failed to create message logs file", zap.Error(err))
		return nil, err
	}

	fileOpts := chroom.ChatRoomFileOptions{
		Writer: w,
		File:   f,
	}
	cr, err := chroom.JoinChatRoom(ctx, h.logger, h.pubsub, h.host.ID(), room, nick, fileOpts)
	if err != nil {
		h.logger.Error("failed to join to room", zap.Error(err), zap.String("room", room))
		return nil, err
	}

	h.logger.Info("user joined to room")

	h.cr[usr] = cr
	return cr, nil
}

func (h *Handler) SendMessage(ctx context.Context, room, nick, filename string, message []byte) {
	logger := h.logger.With(zap.String("room", room), zap.String("nick", nick), zap.String("filename", filename))

	usr := User{Room: room, Nick: nick}
	cr, ok := h.cr[usr]
	if !ok {
		h.JoinRoom(ctx, room, nick) // again connect if not found
	}

	cr, ok = h.cr[usr]
	if !ok {
		logger.Error("chat room not founded or was unsubscribed")
	}

	switch filename {
	case "":
		cr.SendMessage(ctx, logger, nick, string(message))
	default:
		cr.SendMessageWithFile(ctx, logger, nick, filename, message)
	}
}

func (h *Handler) Clear() error {
	for _, cr := range h.cr {
		if err := cr.Close(); err != nil {
			return err
		}
	}
	h.cr = make(map[User]*chroom.ChatRoom)
	return nil
}

func (h *Handler) Shutdown() error {
	for _, cr := range h.cr {
		if err := cr.Close(); err != nil {
			return err
		}
	}

	if err := h.host.Close(); err != nil {
		return err
	}
	return nil
}

func messageLogWritter(room string) (*os.File, *bufio.Writer, error) {
	filepath := fmt.Sprintf("%s/%s.log", pkg.MsgDir, room)
	logFile, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to open room log file")
	}

	writer := bufio.NewWriter(logFile)
	return logFile, writer, nil
}
