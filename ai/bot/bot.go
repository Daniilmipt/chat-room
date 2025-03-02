package bot

import (
	"bytes"
	"chatroom/chat/api"
	"chatroom/chat/chatroom"
	apiconfig "chatroom/chat/config"
	"chatroom/config"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const defaultFileName = ""

type BotService struct {
	cfg    config.ApiConfig
	logger *zap.Logger
	api    *api.Handler

	room string
	nick string
}

type User struct {
	Room string
	Nick string
}

func NewBotService(cfg config.ApiConfig, logger *zap.Logger, usr User) *BotService {
	loggerNew := logger.With(zap.String("room", usr.Room), zap.String("nick", usr.Nick))

	return &BotService{
		cfg:    cfg,
		logger: loggerNew,
		room:   usr.Room,
		nick:   usr.Nick,
	}
}

func (s *BotService) Run(ctx context.Context) error {
	modelHost, _ := strconv.Atoi(s.cfg.Port)
	modelHost++

	api := api.NewHandler(ctx, s.logger, apiconfig.Config{Host: s.cfg.Host, Port: strconv.Itoa(modelHost)})

	s.logger.Info("bot started joined to room")
	defer s.logger.Info("bot finished joined to room")

	cr, err := api.JoinRoom(ctx, s.room, s.nick)
	if err != nil {
		return errors.Wrap(err, "failed to connect bot to chat")
	}

	usrMsgCh := s.readLoop(ctx, cr)
	go func() {
		for msg := range usrMsgCh {
			modelMsg, err := getModelMessage(string(msg))
			if err != nil {
				s.logger.Error("failed to get message from ai model", zap.Error(err))
				continue
			}

			cr.SendMessage(ctx, s.logger, s.nick, modelMsg)
		}

		api.Shutdown() // chtroom cr will be closed inside
	}()

	return nil
}

func getModelMessage(msg string) (string, error) {
	type Request struct {
		Message string `json:"message"`
	}

	requestBody, _ := json.Marshal(Request{Message: msg})
	resp, err := http.Post("http://localhost:5000/chat", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", errors.Wrap(err, "failed to API request for ai model")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		return "", errors.New("empty response from model API")
	}

	type Response struct {
		Response string `json:"response"`
	}
	response := new(Response)
	if err := json.Unmarshal(body, response); err != nil {
		return "", errors.Wrap(err, "failed to unmarshall API response from ai model")
	}

	return response.Response, nil
}

func (s *BotService) readLoop(ctx context.Context, cr *chatroom.ChatRoom) <-chan []byte {
	usrMsgCh := make(chan []byte, 1)

	go func() {
		defer close(usrMsgCh)
		for {
			msg, err := cr.Sub.Next(ctx)
			if err != nil { // TODO error "subrscription canceled"
				s.logger.Error("failed to read next message in room", zap.Error(err))
				return
			}

			if msg == nil {
				s.logger.Error("get nil message from topic", zap.Any("chat_room", cr))
				return
			}

			cm := &chatroom.ChatMessage{}
			if err = json.Unmarshal(msg.Data, cm); err != nil {
				s.logger.Error("failed to unmarshal message in room", zap.Error(err), zap.ByteString("message", msg.Data))
				continue
			}

			if cm.FileName != "" { // skip files and not text messages
				continue
			}

			if msg.ReceivedFrom == cr.Self {
				continue
			}

			usrMsgCh <- cm.Message
		}
	}()

	return usrMsgCh
}
