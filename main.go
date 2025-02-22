package main

import (
	"chatroom/config"
	"chatroom/utils"
	"chatroom/models"
	"chatroom/router"
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	cfg := config.ParseConfig()

	logger, f := utils.SetupLogger()
	defer f.Close()

	// messages from web
	msgCh := make(chan models.Message, 100)
	defer close(msgCh)

	ctx := context.Background()

	// server
	r := router.Init(ctx, cfg, logger, msgCh)
	r.Run()

	// shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	logger.Info("Shutting down server...")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := r.Shutdown(ctx); err != nil {
			logger.Error("failed to shutdown server", zap.Error(err))
		}
		logger.Info("Server, peer and host exited gracefully")
	}()
	wg.Wait()

}
