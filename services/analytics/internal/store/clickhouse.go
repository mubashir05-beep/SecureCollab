package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseAnalyticsStore struct {
	db *sql.DB
}

func NewClickHouseAnalyticsStore(dsn string) (*ClickHouseAnalyticsStore, error) {
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping clickhouse: %w", err)
	}
	return &ClickHouseAnalyticsStore{db: db}, nil
}

func (s *ClickHouseAnalyticsStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *ClickHouseAnalyticsStore) GetMessageVolume(ctx context.Context, windowHours int) (MessageVolume, error) {
	if windowHours <= 0 {
		windowHours = 24
	}

	query := `
		SELECT
			count() AS total_messages,
			countIf(created_at >= now() - toIntervalHour(?)) AS messages_in_window
		FROM encrypted_messages_analytics
	`

	var total int64
	var inWindow int64
	if err := s.db.QueryRowContext(ctx, query, windowHours).Scan(&total, &inWindow); err != nil {
		return MessageVolume{}, fmt.Errorf("query clickhouse volume: %w", err)
	}

	return MessageVolume{
		TotalMessages:   total,
		MessagesLast24h: inWindow,
		WindowHours:     windowHours,
	}, nil
}

// NewAnalyticsStoreWithClickHouse tries ClickHouse first, then falls back to Postgres, then in-memory.
func NewAnalyticsStoreWithClickHouse() (AnalyticsStore, func() error, error) {
	clickhouseDSN := strings.TrimSpace(os.Getenv("CLICKHOUSE_DSN"))
	if clickhouseDSN != "" {
		chStore, err := NewClickHouseAnalyticsStore(clickhouseDSN)
		if err == nil {
			return chStore, chStore.Close, nil
		}
		// Fall through to Postgres/in-memory if ClickHouse unavailable
	}

	return NewAnalyticsStoreFromEnv()
}
