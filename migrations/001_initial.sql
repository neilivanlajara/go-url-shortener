CREATE TABLE IF NOT EXISTS urls (
    id         BIGSERIAL PRIMARY KEY,
    shortcode  VARCHAR(20)  UNIQUE NOT NULL,
    long_url   TEXT         NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_urls_shortcode ON urls(shortcode);

CREATE TABLE IF NOT EXISTS clicks (
    id         BIGSERIAL PRIMARY KEY,
    shortcode  VARCHAR(20)  NOT NULL REFERENCES urls(shortcode),
    clicked_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    user_agent TEXT,
    ip_address INET
);

CREATE INDEX IF NOT EXISTS idx_clicks_shortcode ON clicks(shortcode);
