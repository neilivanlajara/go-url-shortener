package handler

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"go-url-shortener/internal/cache"
	"go-url-shortener/internal/repository"
)

const shortcodeLen = 6
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const cacheTTL = 24 * time.Hour

type URLHandler struct {
	db      *pgxpool.Pool
	cache   *redis.Client
	baseURL string
}

func NewURLHandler(db *pgxpool.Pool, cache *redis.Client, baseURL string) *URLHandler {
	return &URLHandler{db: db, cache: cache, baseURL: baseURL}
}

// POST /shorten
// body: {"url": "https://example.com"}
func (h *URLHandler) Shorten(c *gin.Context) {
	var body struct {
		URL string `json:"url" binding:"required,url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or missing url"})
		return
	}

	shortcode, err := generateShortcode()
	if err != nil {
		slog.Error("shortcode generation failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	u, err := repository.SaveURL(c.Request.Context(), h.db, shortcode, body.URL)
	if err != nil {
		slog.Error("save url failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save url"})
		return
	}

	cache.Set(c.Request.Context(), h.cache, shortcode, body.URL, cacheTTL)

	c.JSON(http.StatusCreated, gin.H{
		"shortcode": u.Shortcode,
		"short_url": h.baseURL + "/" + u.Shortcode,
		"long_url":  u.LongURL,
	})
}

// GET /:shortcode
func (h *URLHandler) Redirect(c *gin.Context) {
	shortcode := c.Param("shortcode")
	ctx := c.Request.Context()

	// Cache hit
	if longURL, err := cache.Get(ctx, h.cache, shortcode); err == nil {
		go trackClick(h.db, shortcode, c.GetHeader("User-Agent"), c.ClientIP())
		c.Redirect(http.StatusFound, longURL)
		return
	}

	// Cache miss → DB
	u, err := repository.GetByShortcode(ctx, h.db, shortcode)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "shortcode not found"})
		return
	}
	if err != nil {
		slog.Error("get url failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if u.DeletedAt != nil {
		c.JSON(http.StatusGone, gin.H{"error": "url has been deleted"})
		return
	}
	if repository.IsExpired(u) {
		c.JSON(http.StatusGone, gin.H{"error": "url has expired"})
		return
	}

	cache.Set(ctx, h.cache, shortcode, u.LongURL, cacheTTL)
	go trackClick(h.db, shortcode, c.GetHeader("User-Agent"), c.ClientIP())
	c.Redirect(http.StatusFound, u.LongURL)
}

// DELETE /:shortcode
func (h *URLHandler) Delete(c *gin.Context) {
	shortcode := c.Param("shortcode")
	ctx := c.Request.Context()

	if err := repository.SoftDelete(ctx, h.db, shortcode); err != nil {
		slog.Error("soft delete failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	cache.Delete(ctx, h.cache, shortcode)
	c.Status(http.StatusNoContent)
}

// GET /:shortcode/stats
func (h *URLHandler) Stats(c *gin.Context) {
	shortcode := c.Param("shortcode")
	ctx := c.Request.Context()

	u, err := repository.GetByShortcode(ctx, h.db, shortcode)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "shortcode not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	clicks, err := repository.GetClicks(ctx, h.db, shortcode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"shortcode":   u.Shortcode,
		"long_url":    u.LongURL,
		"created_at":  u.CreatedAt,
		"deleted":     u.DeletedAt != nil,
		"click_count": len(clicks),
		"clicks":      clicks,
	})
}

// GET /urls
func (h *URLHandler) ListAll(c *gin.Context) {
	urls, err := repository.GetAllURLs(c.Request.Context(), h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, urls)
}

func trackClick(db *pgxpool.Pool, shortcode, userAgent, ip string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := repository.SaveClick(ctx, db, shortcode, userAgent, ip); err != nil {
		slog.Error("track click failed", "shortcode", shortcode, "error", err)
	}
}

func generateShortcode() (string, error) {
	b := make([]byte, shortcodeLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}
