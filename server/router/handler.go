package router

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"social/pkg"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

const maxWorkers = 10

//go:embed index.html view.html room.html
var content embed.FS

type ChatHandler struct {
	logger *zap.Logger
	msgCh  chan string
}

func NewChatHandler(logger *zap.Logger, msgCh chan string) *ChatHandler {
	logger = logger.With(zap.String("id", uuid.New().String()))

	return &ChatHandler{logger: logger, msgCh: msgCh}
}

func (h *ChatHandler) joinToRoom(c *gin.Context) (int, error) {
	room := c.Query("room")
	nick := c.Query("nick")

	if room == "" || nick == "" {
		return http.StatusBadRequest, errors.New("missing room or nick")
	}

	cmd := exec.Command("/home/daniil/blockchain/social_network/chat/chat", "-nick="+nick, "-room="+room)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error getting StdinPipe: %s", err)
	}

	go func() {
		fmt.Println("start read channel")
		for msg := range h.msgCh {
			fmt.Printf("get from channel %s/n", msg)
			if _, err := stdin.Write([]byte(msg + "\n")); err != nil {
				h.logger.Error("fail to write in stdin", zap.Error(err))
				return
			}
		}
		if err := stdin.Close(); err != nil {
			h.logger.Error("fail to close stdin", zap.Error(err))
			return
		}
	}()

	if err := cmd.Start(); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error starting command: %s", err)
	}
	h.logger.Info("join to chat room", zap.String("room", room), zap.String("nick", nick))

	return http.StatusOK, nil
}

func (h *ChatHandler) GetRoomView(c *gin.Context) {
	code, err := h.joinToRoom(c)
	if err != nil {
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	data, err := content.ReadFile("view.html")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

func (h *ChatHandler) GetRoomsListView(c *gin.Context) {
	data, err := content.ReadFile("room.html")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

func (h *ChatHandler) GetAuthView(c *gin.Context) {
	data, err := content.ReadFile("index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "error loading page")
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
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
	var request struct {
		Room    string `json:"room"`
		Nick    string `json:"nick"`
		Message string `json:"message"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("invalid request: %s", err).Error()})
		return
	}
	fmt.Println(request)

	if request.Room == "" || request.Message == "" || request.Nick == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing room, room, or message"})
		return
	}

	fmt.Printf("send to channel %s/n", request.Message)
	h.msgCh <- request.Message
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

func (h *ChatHandler) iterateByMessageFiles(files []fs.DirEntry) map[string]string {
	ctx := context.Background()
	sem := semaphore.NewWeighted(int64(maxWorkers))

	var directory []string
	rooms := make(map[string]string)
	wg := &sync.WaitGroup{}
	for _, f := range files {
		if err := sem.Acquire(ctx, 1); err != nil {
			h.logger.Error("failed semaphore when read last message", zap.Error(err))
			continue
		}
		wg.Add(1)
		go func() {
			defer sem.Release(1)
			defer wg.Done()
			if f.IsDir() || filepath.Ext(f.Name()) != ".log" {
				directory = append(directory, f.Name())
				return
			}

			filePath := "../messages/" + f.Name()
			lastLine, err := pkg.GetLastLine(filePath)
			if err != nil {
				h.logger.Error("error reading last line of file", zap.String("file", f.Name()), zap.Error(err))
				return
			}

			filename := strings.TrimSuffix(f.Name(), ".log")
			rooms[filename] = lastLine
		}()
	}
	wg.Wait()

	if len(directory) > 0 {
		h.logger.Warn("unexpected directories or other file extensions in messages folder",
			zap.Strings("directories", directory),
		)
	}
	return rooms
}
