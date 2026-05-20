package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db    *pgxpool.Pool
	cache *redis.Client
}

func NewHealthHandler(db *pgxpool.Pool, cache *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, cache: cache}
}

func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	resp := gin.H{"status": "ok", "postgres": "ok", "redis": "ok"}
	code := http.StatusOK

	if err := h.db.Ping(ctx); err != nil {
		resp["postgres"] = "unavailable"
		resp["status"] = "degraded"
		code = http.StatusServiceUnavailable
	}

	if err := h.cache.Ping(ctx).Err(); err != nil {
		resp["redis"] = "unavailable"
		resp["status"] = "degraded"
		code = http.StatusServiceUnavailable
	}

	c.JSON(code, resp)
}
