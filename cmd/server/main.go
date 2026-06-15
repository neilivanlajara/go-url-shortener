package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go-url-shortener/internal/cache"
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/handler"
	"go-url-shortener/internal/repository"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()
	ctx := context.Background()

	db, err := repository.NewPool(ctx, cfg.DBURL)
	if err != nil {
		slog.Error("postgres connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("postgres connected")

	redisClient, err := cache.NewClient(ctx, cfg.RedisURL)
	if err != nil {
		slog.Error("redis connection failed", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	slog.Info("redis connected")

	r := gin.New()
	r.Use(gin.Recovery())

	healthHandler := handler.NewHealthHandler(db, redisClient)
	urlHandler := handler.NewURLHandler(db, redisClient, cfg.BaseURL)

	r.GET("/health", healthHandler.Check)

	r.POST("/shorten", urlHandler.Shorten)
	r.GET("/urls", urlHandler.ListAll)
	r.GET("/:shortcode", urlHandler.Redirect)
	r.GET("/:shortcode/stats", urlHandler.Stats)
	r.DELETE("/:shortcode", urlHandler.Delete)

	slog.Info("server starting", "port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
