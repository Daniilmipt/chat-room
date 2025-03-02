package handler

import (
	"chatroom/ai/bot"
	"chatroom/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type BotHandler struct {
	logger *zap.Logger
	cfg    config.ApiConfig
}

func NewBotHandler(cfg config.ApiConfig, logger *zap.Logger) *BotHandler {
	return &BotHandler{
		logger: logger,
		cfg:    cfg,
	}
}

func (h *BotHandler) CreateBot(c *gin.Context) (int, error) {
	type RequestBot struct {
		BotNick string `json:"botNick"`
		Room    string `json:"room"`
	}

	var request RequestBot
	if err := c.BindJSON(&request); err != nil {
		h.logger.Error("failed to bind json request", zap.Error(err))
		return http.StatusBadRequest, errors.Wrap(err, "invalid request")
	}

	usr := bot.User{Room: request.Room, Nick: request.BotNick}
	botService := bot.NewBotService(h.cfg, h.logger, usr)
	if err := botService.Run(c); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
