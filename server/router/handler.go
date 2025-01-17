package router

import (
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"social/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:embed index.html room.html room_list.html
var content embed.FS

type ChatHandler struct {
	logger   *zap.Logger
	msgCh    chan models.Messages
	stdinMap map[string]*io.WriteCloser
}

func NewChatHandler(logger *zap.Logger, msgCh chan models.Messages) *ChatHandler {
	logger = logger.With(zap.String("id", uuid.New().String()))

	return &ChatHandler{logger: logger, msgCh: msgCh, stdinMap: make(map[string]*io.WriteCloser)}
}

func (h *ChatHandler) GetRoomsLastMessage(c *gin.Context) {
	files, err := os.ReadDir("../messages")
	if err != nil {
		h.logger.Error("error reading directory with messages", zap.Error(err))
		c.String(http.StatusInternalServerError, fmt.Errorf("error reading directory with messages: %s", err).Error())
		return
	}

	rooms := h.iterateByMessageFiles(files)
	c.JSON(http.StatusOK, rooms)
}

func (h *ChatHandler) GetMessagesFile(c *gin.Context) {
	room := c.Query("room")
	if room == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing room"})
		return
	}

	logFilePath := fmt.Sprintf("../messages/%s.log", room)
	if _, err := os.Stat(logFilePath); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Errorf("room log file not found: %q", err).Error()})
		return
	}

	c.File(logFilePath)
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	var request models.Messages
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("invalid request: %s", err).Error()})
		return
	}

	if !request.Validate() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing room, room, or message"})
		return
	}

	h.msgCh <- request
}
