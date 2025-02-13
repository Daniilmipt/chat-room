package router

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *ChatHandler) GetRoomView(c *gin.Context) {
	go h.sendMessageInOut(c)

	room := c.Query("room")
	nick := c.Query("nick")

	if room == "" || nick == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("missing room or nick").Error()})
	}

	if err := h.joinToRoom(c, room); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data, err := content.ReadFile("room.html")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

func (h *ChatHandler) GetRoomsListView(c *gin.Context) {
	data, err := content.ReadFile("room_list.html")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

func (h *ChatHandler) GetAuthView(c *gin.Context) {
	data, err := content.ReadFile("login.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "error loading page")
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}
