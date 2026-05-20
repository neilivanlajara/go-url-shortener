package model

import "time"

type URL struct {
	ID        int64      `json:"id"`
	Shortcode string     `json:"shortcode"`
	LongURL   string     `json:"long_url"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	DeletedAt *time.Time `json:"-"`
}

type Click struct {
	ID        int64     `json:"id"`
	Shortcode string    `json:"shortcode"`
	ClickedAt time.Time `json:"clicked_at"`
	UserAgent string    `json:"user_agent,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
}
