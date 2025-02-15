package main

import (
	"chatroom/config"
	"chatroom/pkg"
	"chatroom/pkg/models"
	"chatroom/router"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	cfg := config.ParseConfig()

	logger, f := pkg.SetupLogger()
	defer f.Close()

	// messages from frontend
	msgCh := make(chan models.Message, 100)
	defer close(msgCh)

	// server
	r := router.Init(cfg, logger, msgCh)
	r.Run()
	logger.Info(fmt.Sprintf("Server started at http://localhost:%s", cfg.Frontend.Port))

	// messages dir storage
	if err := os.Mkdir("./messages", os.ModePerm); err != nil {
		logger.Info("can not create messages dis", zap.Error(err))
	}

	// shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := r.Shutdown(ctx); err != nil {
			logger.Error("failed to shutdown server", zap.Error(err))
		}
		logger.Info("Server exited gracefully")
	}()
	wg.Wait()

}
