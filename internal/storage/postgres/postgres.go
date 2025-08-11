package postgres

import (
	"analiticsURLShortener/internal/config"
	"analiticsURLShortener/internal/storage"
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func InitDB(cfg *config.Config) (*Storage, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("db connection error: %v", err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("couldn't connect to the DB: %v", err)
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave, alias string) (int64, error) {
	var id int64
	err := s.db.QueryRow("INSERT INTO url (url, alias) VALUES ($1, $2) RETURNING id", urlToSave, alias).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("couldn't insert URL: %v", err)
	}
	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	var url string
	err := s.db.QueryRow("SELECT url FROM url WHERE alias = $1", alias).Scan(&url)
	if err != nil {
		return "", fmt.Errorf("couldn't get URL: %v", err)
	}
	return url, nil
}

func (s *Storage) SaveAnalytics(alias string, userAgent string) error {
	var urlID int
	err := s.db.QueryRow("SELECT id FROM url WHERE alias = $1", alias).Scan(&urlID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrURLNotFound
		}
		return fmt.Errorf("couldn't get url id: %w", err)
	}

	_, err = s.db.Exec("INSERT INTO url_analytics (url_id, user_agent) VALUES ($1, $2)", urlID, userAgent)
	if err != nil {
		return fmt.Errorf("couldn't save analytics: %w", err)
	}

	return nil
}

func (s *Storage) GetAnalytics(alias string) (storage.AnalyticsData, error) {
	var urlID int
	err := s.db.QueryRow("SELECT id FROM url WHERE alias = $1", alias).Scan(&urlID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.AnalyticsData{}, storage.ErrURLNotFound
		}
		return storage.AnalyticsData{}, fmt.Errorf("couldn't get url id: %w", err)
	}

	var totalClicks int64
	err = s.db.QueryRow("SELECT COUNT(*) FROM url_analytics WHERE url_id = $1", urlID).Scan(&totalClicks)
	if err != nil {
		return storage.AnalyticsData{}, fmt.Errorf("couldn't get total clicks: %w", err)
	}

	userAgentCounts := make(map[string]int64)
	rows, err := s.db.Query("SELECT user_agent, COUNT(*) FROM url_analytics WHERE url_id = $1 GROUP BY user_agent", urlID)
	if err != nil {
		return storage.AnalyticsData{}, fmt.Errorf("couldn't get user agent stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userAgent string
		var count int64
		if err := rows.Scan(&userAgent, &count); err != nil {
			return storage.AnalyticsData{}, fmt.Errorf("couldn't scan user agent row: %w", err)
		}
		userAgentCounts[userAgent] = count
	}

	dailyCounts := make(map[string]int64)
	rows, err = s.db.Query("SELECT DATE(created_at), COUNT(*) FROM url_analytics WHERE url_id = $1 GROUP BY DATE(created_at)", urlID)
	if err != nil {
		return storage.AnalyticsData{}, fmt.Errorf("couldn't get daily stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var date string
		var count int64
		if err := rows.Scan(&date, &count); err != nil {
			return storage.AnalyticsData{}, fmt.Errorf("couldn't scan daily row: %w", err)
		}
		dailyCounts[date] = count
	}

	monthlyCounts := make(map[string]int64)
	rows, err = s.db.Query("SELECT to_char(created_at, 'YYYY-MM'), COUNT(*) FROM url_analytics WHERE url_id = $1 GROUP BY to_char(created_at, 'YYYY-MM')", urlID)
	if err != nil {
		return storage.AnalyticsData{}, fmt.Errorf("couldn't get monthly stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var month string
		var count int64
		if err := rows.Scan(&month, &count); err != nil {
			return storage.AnalyticsData{}, fmt.Errorf("couldn't scan monthly row: %w", err)
		}
		monthlyCounts[month] = count
	}

	return storage.AnalyticsData{
		TotalClicks: totalClicks,
		UserAgents:  userAgentCounts,
		Daily:       dailyCounts,
		Monthly:     monthlyCounts,
	}, nil
}
