package server

import (
	"go-pubsub-api/internal/config"
	"go-pubsub-api/internal/handlers"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()
	handler := handlers.NewHandler()

	hub := handlers.NewHub()
	go hub.Run()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     config.Cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	r.GET("/", s.HelloWorldHandler)
	r.POST("/publish", handler.PublishHandler(hub))
	r.GET("/subscribe/:topic", handler.SubscribeHandler(hub))
	r.POST("/auth/callback", handler.AuthCallbackHandler)

	ui := r.Group("/ui")
	{
		ui.GET("/dashboard", handlers.DashboardHandler)
	}
	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}
