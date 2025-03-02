package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *ChatHandler) GetRoomView(c *gin.Context) (int, []byte, error) {
	go h.sendMessageInOut(c)

	room := c.Query("room")
	nick := c.Query("nick")

	if room == "" || nick == "" {
		h.logger.Error("empty room or nick in request", zap.String("room", room), zap.String("nick", nick))
		return http.StatusBadRequest, nil, errors.New("missing room or nick")
	}

	cr, err := h.api.JoinRoom(c, room, nick)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	go cr.ReadLoop(c, h.logger)

	data, err := content.ReadFile("room.html")
	if err != nil {
		h.logger.Error("can not read room.html", zap.Error(err))
		return http.StatusInternalServerError, nil, err
	}

	return http.StatusOK, data, nil
}

func (h *ChatHandler) GetRoomsListView(c *gin.Context) (int, []byte, error) {
	data, err := content.ReadFile("room_list.html")
	if err != nil {
		h.logger.Error("can not read room_list.html", zap.Error(err))
		return http.StatusInternalServerError, nil, err
	}

	return http.StatusOK, data, nil
}

func (h *ChatHandler) GetAuthView(c *gin.Context) (int, []byte, error) {
	data, err := content.ReadFile("login.html")
	if err != nil {
		h.logger.Error("can not read login.html", zap.Error(err))
		return http.StatusInternalServerError, nil, err
	}

	return http.StatusOK, data, nil
}
