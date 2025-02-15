package handler

import (
	"chatroom/chat/api"
	apiconfig "chatroom/chat/config"
	"chatroom/config"
	"chatroom/pkg/models"
	"embed"
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:embed login.html room.html room_list.html
var content embed.FS

type ChatHandler struct {
	logger *zap.Logger
	msgCh  chan models.Message

	api api.Handler
}

func NewChatHandler(cfg config.BackendConfig, logger *zap.Logger, msgCh chan models.Message) *ChatHandler {
	logger = logger.With(zap.String("id", uuid.New().String()))
	api := api.NewHandler(logger, apiconfig.Config{Host: cfg.Host, Port: cfg.Port})

	return &ChatHandler{
		logger: logger,
		msgCh:  msgCh,
		api:    api,
	}
}

func (h *ChatHandler) GetRoomsLastMessage(c *gin.Context) (int, map[string]string, error) {
	files, err := os.ReadDir("./messages")
	if err != nil {
		h.logger.Error("error reading directory with messages", zap.Error(err))
		return http.StatusInternalServerError, nil, errors.Wrap(err, "error reading directory with messages")
		// c.String(http.StatusInternalServerError, fmt.Errorf("error reading directory with messages: %s", err).Error())
		// return
	}

	rooms := h.iterateByMessageFiles(files)
	return http.StatusOK, rooms, errors.Wrap(err, "error iterate by messages")
	// c.JSON(http.StatusOK, rooms)
}

func (h *ChatHandler) GetMessagesFile(c *gin.Context) (int, string, error) {
	room := c.Query("room")
	if room == "" {
		return http.StatusBadRequest, "", errors.New("empty room")
		// c.JSON(http.StatusBadRequest, gin.H{"error": "missing room"})
		// return
	}

	logFilePath := fmt.Sprintf("./messages/%s.log", room)
	if _, err := os.Stat(logFilePath); err != nil {
		return http.StatusNotFound, "", errors.Wrap(err, "room log file not found")
		// c.JSON(http.StatusNotFound, gin.H{"error": fmt.Errorf("room log file not found: %q", err).Error()})
		// return
	}

	return http.StatusOK, logFilePath, nil
	// c.File(logFilePath)
}

func (h *ChatHandler) SendMessage(c *gin.Context) (int, error) {
	var request models.MessageRequest
	if err := c.BindJSON(&request); err != nil {
		h.logger.Error("failed to bind json request", zap.Error(err))
		// c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("invalid request: %s", err).Error()})
		return http.StatusBadRequest, errors.Wrap(err, "invalid request")
	}

	if !request.Validate() {
		h.logger.Error("failed to validate request", zap.Any("request", request))
		return http.StatusBadRequest, errors.New("missing room, nick or message")

		// c.JSON(http.StatusBadRequest, gin.H{"error": "missing room, nick, or message"})
		// return
	}

	message, err := request.ToMessage()
	if err != nil {
		h.logger.Error("failed to convert request to message", zap.Error(err), zap.Any("request", request))
		return http.StatusBadRequest, errors.New("failed to convert request to message")
	}

	h.msgCh <- message
	return http.StatusOK, nil
}

func (h *ChatHandler) LogOut(c *gin.Context) (int, error) {
	h.api.Clear()
	return http.StatusOK, nil
}
