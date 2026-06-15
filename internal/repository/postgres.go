package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go-url-shortener/internal/model"
)

func NewPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func SaveURL(ctx context.Context, pool *pgxpool.Pool, shortcode, longURL string) (*model.URL, error) {
	var u model.URL
	err := pool.QueryRow(ctx, `
		INSERT INTO urls (shortcode, long_url)
		VALUES ($1, $2)
		RETURNING id, shortcode, long_url, created_at, expires_at, deleted_at
	`, shortcode, longURL).Scan(
		&u.ID, &u.Shortcode, &u.LongURL, &u.CreatedAt, &u.ExpiresAt, &u.DeletedAt,
	)
	return &u, err
}

func GetByShortcode(ctx context.Context, pool *pgxpool.Pool, shortcode string) (*model.URL, error) {
	var u model.URL
	err := pool.QueryRow(ctx, `
		SELECT id, shortcode, long_url, created_at, expires_at, deleted_at
		FROM urls
		WHERE shortcode = $1
	`, shortcode).Scan(
		&u.ID, &u.Shortcode, &u.LongURL, &u.CreatedAt, &u.ExpiresAt, &u.DeletedAt,
	)
	return &u, err
}

func SoftDelete(ctx context.Context, pool *pgxpool.Pool, shortcode string) error {
	_, err := pool.Exec(ctx, `
		UPDATE urls SET deleted_at = NOW() WHERE shortcode = $1 AND deleted_at IS NULL
	`, shortcode)
	return err
}

func SaveClick(ctx context.Context, pool *pgxpool.Pool, shortcode, userAgent, ip string) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO clicks (shortcode, user_agent, ip_address)
		VALUES ($1, $2, $3)
	`, shortcode, userAgent, ip)
	return err
}

func GetClicks(ctx context.Context, pool *pgxpool.Pool, shortcode string) ([]model.Click, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, shortcode, clicked_at, user_agent, ip_address
		FROM clicks
		WHERE shortcode = $1
		ORDER BY clicked_at DESC
	`, shortcode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clicks []model.Click
	for rows.Next() {
		var c model.Click
		if err := rows.Scan(&c.ID, &c.Shortcode, &c.ClickedAt, &c.UserAgent, &c.IPAddress); err != nil {
			return nil, err
		}
		clicks = append(clicks, c)
	}
	return clicks, rows.Err()
}

func GetAllURLs(ctx context.Context, pool *pgxpool.Pool) ([]model.URL, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, shortcode, long_url, created_at, expires_at, deleted_at
		FROM urls
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []model.URL
	for rows.Next() {
		var u model.URL
		if err := rows.Scan(&u.ID, &u.Shortcode, &u.LongURL, &u.CreatedAt, &u.ExpiresAt, &u.DeletedAt); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, rows.Err()
}

func IsExpired(u *model.URL) bool {
	return u.ExpiresAt != nil && u.ExpiresAt.Before(time.Now())
}
