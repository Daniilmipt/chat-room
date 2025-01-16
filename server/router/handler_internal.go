package router

import (
	"context"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"runtime"
	"social/pkg"
	"strings"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

const maxWorkersFileIter = 10

func (h *ChatHandler) sendMessageInOut() {
	fmt.Println("start read msg from channel")

	for msg := range h.msgCh {
		fmt.Println("read message from channel", msg)

		stdin := h.stdinMap[msg.Room]
		if stdin == nil {
			if err := h.joinToRoom(msg.Room, msg.Nick); err != nil {
				h.logger.Error("fail to join to room",
					zap.Error(err),
					zap.Any("room-message", msg),
				)
			}
			stdin = h.stdinMap[msg.Room]
		}

		if _, err := (*stdin).Write([]byte(msg.Message + "\n")); err != nil {
			h.logger.Error("fail to write message to chat",
				zap.Error(err),
				zap.Any("room-message", msg),
			)
		}
	}
}

func getFileExecPath() string {
	var executable string
	switch runtime.GOOS {
	case "windows":
		executable = "chat.exe"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			executable = "chat_mac_arm"
		} else {
			executable = "chat_mac"
		}
	case "linux":
		executable = "chat_linux"
	default:
		panic("Unsupported operating system")
	}
	return "../chat/" + executable
}

func (h *ChatHandler) joinToRoom(room, nick string) error {
	execPath := getFileExecPath()
	h.logger.Info("start chat process", zap.String("path", execPath))
	cmd := exec.Command(execPath, "-nick="+nick, "-room="+room)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error getting StdinPipe: %s", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %s", err)
	}

	h.stdinMap[room] = &stdin
	h.logger.Info("join to chat room", zap.String("room", room), zap.String("nick", nick))
	return nil
}

func (h *ChatHandler) iterateByMessageFiles(files []fs.DirEntry) map[string]string {
	ctx := context.Background()
	sem := semaphore.NewWeighted(int64(maxWorkersFileIter))

	var directory []string
	rooms := make(map[string]string)
	wg := &sync.WaitGroup{}
	for _, f := range files {
		if err := sem.Acquire(ctx, 1); err != nil {
			h.logger.Error("failed semaphore when read last message", zap.Error(err))
			continue
		}
		wg.Add(1)
		go func(f fs.DirEntry) {
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
		}(f)
	}
	wg.Wait()

	if len(directory) > 0 {
		h.logger.Warn("unexpected directories or other file extensions in messages folder",
			zap.Strings("directories", directory),
		)
	}
	return rooms
}
