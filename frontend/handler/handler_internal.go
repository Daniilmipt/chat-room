package handler

import (
	"chatroom/pkg"
	"context"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

const maxWorkersFileIter = 10

func (h *ChatHandler) sendMessageInOut(ctx context.Context) {
	for msg := range h.msgCh {
		h.api.SendMessage(ctx, msg.Room, msg.Nick, msg.Message)
	}
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

			filePath := "./messages/" + f.Name()
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
