package api

import (
	"bufio"
	"chatroom/chat/config"
	chroom "chatroom/chat/chatroom"
	"chatroom/pkg"
	"chatroom/chat/service"
	"context"
	"fmt"
	"os"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Params struct {
	nick string
	room string
	host string
	port string
}

type Handler struct {
	cr map[string]*chroom.ChatRoom

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
		cr:     make(map[string]*chroom.ChatRoom),
	}
}

func (h *Handler) JoinRoom(ctx context.Context, room, nick string) error {
	if _, ok := h.cr[room]; ok {
		return nil
	}

	f, w, err := messageLogWritter(room)
	if err != nil {
		h.logger.Error("failed to create message logs file", zap.Error(err))
		return err
	}

	h.msgFile = f
	h.writer = w

	cr, err := chroom.JoinChatRoom(ctx, h.logger, h.pubsub, h.host.ID(), room, nick, h.writer)
	if err != nil {
		h.logger.Error("failed to join to room", zap.Error(err), zap.String("room", room))
		return err
	}

	h.cr[room] = cr
	return nil
}

func (h *Handler) SendMessage(ctx context.Context, room, nick string, message []byte) {
	logger := h.logger.With(zap.String("room", room), zap.String("nick", nick))

	cr, ok := h.cr[room]
	if !ok {
		h.JoinRoom(ctx, room, nick)
	}

	cr, ok = h.cr[room]
	if !ok {
		logger.Error("chat room not founded or was unsubscribed")
	}

	cr.SendMessage(ctx, h.logger, nick, message)
}

func (h *Handler) Clear() error {
	for _, cr := range h.cr {
		if err := cr.Close(); err != nil {
			return err
		}
	}
	h.cr = make(map[string]*chroom.ChatRoom)
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
