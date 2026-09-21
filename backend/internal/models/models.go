package models

import "time"

type URL struct {
	ID        int       `json:"id"`
	ShortCode string    `json:"short_code"`
	LongURL   string    `json:"long_url"`
	CreatedAt time.Time `json:"created_at"`
}

type ShortenRequest struct {
	LongURL string `json:"long_url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
	ShortCode string `json:"short_code"`
}

type StatsResponse struct {
	ShortCode string    `json:"short_code"`
	LongURL   string    `json:"long_url"`
	Clicks    int       `json:"clicks"`
	CreatedAt time.Time `json:"created_at"`
}

type AnalyticsEntry struct {
	ID        int       `json:"id"`
	Referrer  string    `json:"referrer"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	ClickedAt time.Time `json:"clicked_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
