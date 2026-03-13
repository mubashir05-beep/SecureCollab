package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type MessageVolume struct {
	TotalMessages   int64
	MessagesLast24h int64
	WindowHours     int
}

type AnalyticsStore interface {
	GetMessageVolume(ctx context.Context, windowHours int) (MessageVolume, error)
}

type inMemoryAnalyticsStore struct {
	mu      sync.Mutex
	times   []time.Time
	clockFn func() time.Time
}

func NewInMemoryAnalyticsStore() *inMemoryAnalyticsStore {
	return &inMemoryAnalyticsStore{times: make([]time.Time, 0), clockFn: time.Now}
}

func (s *inMemoryAnalyticsStore) SetMessageTimes(times []time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.times = append([]time.Time(nil), times...)
}

func (s *inMemoryAnalyticsStore) GetMessageVolume(_ context.Context, windowHours int) (MessageVolume, error) {
	if windowHours <= 0 {
		windowHours = 24
	}
	cutoff := s.clockFn().Add(-time.Duration(windowHours) * time.Hour)

	s.mu.Lock()
	defer s.mu.Unlock()

	var inWindow int64
	for _, ts := range s.times {
		if !ts.Before(cutoff) {
			inWindow++
		}
	}

	return MessageVolume{TotalMessages: int64(len(s.times)), MessagesLast24h: inWindow, WindowHours: windowHours}, nil
}

type PostgresAnalyticsStore struct {
	db *sql.DB
}

func NewPostgresAnalyticsStore(databaseURL string) (*PostgresAnalyticsStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &PostgresAnalyticsStore{db: db}, nil
}

func (s *PostgresAnalyticsStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *PostgresAnalyticsStore) GetMessageVolume(ctx context.Context, windowHours int) (MessageVolume, error) {
	if windowHours <= 0 {
		windowHours = 24
	}
	query := `
		SELECT
			COUNT(*) AS total_messages,
			SUM(CASE WHEN created_at >= NOW() - ($1::text || ' hours')::interval THEN 1 ELSE 0 END) AS messages_in_window
		FROM encrypted_messages
	`
	var total int64
	var inWindow int64
	if err := s.db.QueryRowContext(ctx, query, windowHours).Scan(&total, &inWindow); err != nil {
		return MessageVolume{}, fmt.Errorf("query message volume: %w", err)
	}
	return MessageVolume{TotalMessages: total, MessagesLast24h: inWindow, WindowHours: windowHours}, nil
}

func NewAnalyticsStoreFromEnv() (AnalyticsStore, func() error, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		inMem := NewInMemoryAnalyticsStore()
		return inMem, func() error { return nil }, nil
	}

	postgresStore, err := NewPostgresAnalyticsStore(databaseURL)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			return nil, nil, fmt.Errorf("init postgres analytics store: %s (%s)", pgErr.Message, pgErr.Code)
		}
		return nil, nil, err
	}
	return postgresStore, postgresStore.Close, nil
}
