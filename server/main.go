package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"social/pkg"
	"social/pkg/cron"
	"social/pkg/processes"
	chrouter "social/router"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	logger, f := pkg.SetupLogger()
	defer f.Close()

	// schedule
	sch := cron.NewSchedule(
		[]cron.CronJob{
			cron.NewJob(
				logger,
				time.Second*60,
				"kill-chat-processes",
				func() {
					errCh := processes.SheduleProcessCount()
					for err := range errCh {
						logger.Error("fail to kill chat process",
							zap.Error(err),
						)
					}
				},
			),
		},
	)
	sch.Start()

	msgCh := make(chan string, 10)
	defer close(msgCh)

	router := gin.Default()
	router.Use(corsMiddleware())

	chatHandler := chrouter.NewChatHandler(logger, msgCh)

	router.GET("/", chatHandler.GetAuthView)

	authorized := router.Group("/")
	authorized.Use(checkNickMiddleware())
	{
		authorized.GET("/room", chatHandler.GetRoomView)
		authorized.GET("/rooms-list", chatHandler.GetRoomsListView)
		authorized.GET("/rooms-last-message", chatHandler.GetRoomsLastMessage)
		// authorized.GET("/start-chat", chatHandler.JoinChatRoom)
		authorized.GET("/messages", chatHandler.GetMessagesFile)
		authorized.POST("/send-message", chatHandler.SendMessage)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	go func() {
		logger.Info("Server started at http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	<-stop
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wg := sync.WaitGroup{}

	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("failed to shutdown server", zap.Error(err))
		}
		logger.Info("Server exited gracefully")
	}()

	go func() {
		defer wg.Done()
		if err := processes.StopJoinRoom(); err != nil {
			logger.Error("failed to kill \"chat\" process", zap.Error(err))
		}
		logger.Info("Chat process killed gracefully")
	}()
	wg.Wait()

}

func checkNickMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		nick, err := c.Cookie("user")
		if err != nil || nick == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "You must have a 'user' cookie to access this endpoint.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// corsMiddleware adds CORS headers to the response
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, X-Auth-Token")
		}

		// Handle preflight requests
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
