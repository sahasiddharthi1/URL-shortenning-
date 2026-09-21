package store

import (
	"database/sql"
	"fmt"
	"url-shortener/internal/models"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(host, port, user, password, dbname string) (*PostgresStore, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func NewPostgresStoreFromURL(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	store := &PostgresStore{db: db}
	if err := store.migrate(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *PostgresStore) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id          SERIAL PRIMARY KEY,
		short_code  VARCHAR(10) UNIQUE NOT NULL,
		long_url    TEXT NOT NULL,
		created_at  TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS analytics (
		id          SERIAL PRIMARY KEY,
		short_code  VARCHAR(10) REFERENCES urls(short_code),
		referrer    TEXT,
		user_agent  TEXT,
		ip_address  TEXT,
		clicked_at  TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
	CREATE INDEX IF NOT EXISTS idx_analytics_code ON analytics(short_code);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) Close() {
	s.db.Close()
}

func (s *PostgresStore) CreateURL(shortCode, longURL string) (*models.URL, error) {
	var url models.URL
	err := s.db.QueryRow(
		"INSERT INTO urls (short_code, long_url) VALUES ($1, $2) RETURNING id, short_code, long_url, created_at",
		shortCode, longURL,
	).Scan(&url.ID, &url.ShortCode, &url.LongURL, &url.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (s *PostgresStore) GetURLByCode(shortCode string) (*models.URL, error) {
	var url models.URL
	err := s.db.QueryRow(
		"SELECT id, short_code, long_url, created_at FROM urls WHERE short_code = $1",
		shortCode,
	).Scan(&url.ID, &url.ShortCode, &url.LongURL, &url.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (s *PostgresStore) GetURLByLongURL(longURL string) (*models.URL, error) {
	var url models.URL
	err := s.db.QueryRow(
		"SELECT id, short_code, long_url, created_at FROM urls WHERE long_url = $1",
		longURL,
	).Scan(&url.ID, &url.ShortCode, &url.LongURL, &url.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (s *PostgresStore) CodeExists(shortCode string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM urls WHERE short_code = $1",
		shortCode,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *PostgresStore) LogAnalytics(shortCode, referrer, userAgent, ipAddress string) error {
	_, err := s.db.Exec(
		"INSERT INTO analytics (short_code, referrer, user_agent, ip_address) VALUES ($1, $2, $3, $4)",
		shortCode, referrer, userAgent, ipAddress,
	)
	return err
}

func (s *PostgresStore) GetStats(shortCode string) (*models.StatsResponse, []models.AnalyticsEntry, error) {
	var stats models.StatsResponse
	err := s.db.QueryRow(
		"SELECT short_code, long_url, created_at FROM urls WHERE short_code = $1",
		shortCode,
	).Scan(&stats.ShortCode, &stats.LongURL, &stats.CreatedAt)
	if err != nil {
		return nil, nil, err
	}

	err = s.db.QueryRow(
		"SELECT COUNT(*) FROM analytics WHERE short_code = $1",
		shortCode,
	).Scan(&stats.Clicks)
	if err != nil {
		return nil, nil, err
	}

	rows, err := s.db.Query(
		"SELECT id, referrer, user_agent, ip_address, clicked_at FROM analytics WHERE short_code = $1 ORDER BY clicked_at DESC LIMIT 10",
		shortCode,
	)
	if err != nil {
		return &stats, nil, nil
	}
	defer rows.Close()

	var entries []models.AnalyticsEntry
	for rows.Next() {
		var entry models.AnalyticsEntry
		if err := rows.Scan(&entry.ID, &entry.Referrer, &entry.UserAgent, &entry.IPAddress, &entry.ClickedAt); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return &stats, entries, nil
}
