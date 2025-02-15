package router

import (
	"chatroom/config"
	"chatroom/handler"
	"chatroom/pkg/models"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	cfg config.FrontConfig
	g   *gin.Engine
	h   *handler.ChatHandler
}

func Init(cfg config.Config, logger *zap.Logger, msgCh chan models.Message) Router {
	g := gin.Default()
	g.Use(corsMiddleware())

	h := handler.NewChatHandler(cfg.Backend, logger, msgCh)

	r := Router{g: g, h: h, cfg: cfg.Frontend}
	r.setupRouters()

	return r
}

func (r *Router) Run() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", r.cfg.Port),
		Handler: r.g,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
}

func (r *Router) Shutdown(ctx context.Context) error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", r.cfg.Port),
		Handler: r.g,
	}
	return srv.Shutdown(ctx)
}

func (r *Router) setupRouters() {
	r.g.GET("/", handleRequestData(r.h.GetAuthView))

	auth := r.g.Group("/")
	auth.Use(checkNickMiddleware())
	{
		auth.GET("/room", handleRequestData(r.h.GetRoomView))
		auth.GET("/rooms-list", handleRequestData(r.h.GetRoomsListView))
		auth.GET("/rooms-last-message", handleRequestJson(r.h.GetRoomsLastMessage))
		auth.GET("/messages", handleRequestFile(r.h.GetMessagesFile))
		auth.POST("/send-message", handleRequest(r.h.SendMessage))
		auth.GET("/out", handleRequest(r.h.LogOut))
	}
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

type RequestHandlerFunc func(ctx *gin.Context) (int, error)

func handleRequest(handlerFunc RequestHandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		code, err := handlerFunc(c)
		if err != nil {
			c.JSON(code, gin.H{"error": err.Error()})
		}
	}
}

type RequestHandlerDataFunc func(ctx *gin.Context) (int, []byte, error)

func handleRequestData(handlerFunc RequestHandlerDataFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		code, data, err := handlerFunc(c)
		if err != nil {
			c.JSON(code, gin.H{"error": err.Error()})
		}

		c.Data(code, "text/html; charset=utf-8", data)
	}
}

type RequestHandlerJsonResponseFunc func(ctx *gin.Context) (int, map[string]string, error)

func handleRequestJson(handlerFunc RequestHandlerJsonResponseFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		code, data, err := handlerFunc(c)
		if err != nil {
			c.JSON(code, gin.H{"error": err.Error()})
		}

		c.JSON(code, data)
	}
}

type RequestHandlerFileFunc func(ctx *gin.Context) (int, string, error)

func handleRequestFile(handlerFunc RequestHandlerFileFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		code, filePath, err := handlerFunc(c)
		if err != nil {
			c.JSON(code, gin.H{"error": err.Error()})
		}

		c.File(filePath)
	}
}
