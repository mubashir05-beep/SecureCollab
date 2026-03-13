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

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var ErrKeyNotFound = errors.New("public key not found")

type PublicKey struct {
	ID        string
	UserID    string
	KeyType   string
	KeyData   []byte
	CreatedAt time.Time
}

type KeyStore interface {
	SavePublicKey(ctx context.Context, userID, keyType string, keyData []byte) (PublicKey, error)
	GetLatestPublicKey(ctx context.Context, userID, keyType string) (PublicKey, error)
}

type inMemoryKeyStore struct {
	mu   sync.Mutex
	keys map[string][]PublicKey
}

func NewInMemoryKeyStore() KeyStore {
	return &inMemoryKeyStore{keys: make(map[string][]PublicKey)}
}

func (s *inMemoryKeyStore) SavePublicKey(_ context.Context, userID, keyType string, keyData []byte) (PublicKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if keyType == "" {
		keyType = "identity"
	}

	key := PublicKey{
		ID:        uuid.NewString(),
		UserID:    userID,
		KeyType:   keyType,
		KeyData:   append([]byte(nil), keyData...),
		CreatedAt: time.Now().UTC(),
	}

	mapKey := userID + ":" + keyType
	s.keys[mapKey] = append(s.keys[mapKey], key)
	return key, nil
}

func (s *inMemoryKeyStore) GetLatestPublicKey(_ context.Context, userID, keyType string) (PublicKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if keyType == "" {
		keyType = "identity"
	}

	mapKey := userID + ":" + keyType
	entries := s.keys[mapKey]
	if len(entries) == 0 {
		return PublicKey{}, ErrKeyNotFound
	}
	latest := entries[len(entries)-1]
	latest.KeyData = append([]byte(nil), latest.KeyData...)
	return latest, nil
}

type PostgresKeyStore struct {
	db *sql.DB
}

func NewPostgresKeyStore(databaseURL string) (*PostgresKeyStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &PostgresKeyStore{db: db}, nil
}

func (s *PostgresKeyStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *PostgresKeyStore) SavePublicKey(ctx context.Context, userID, keyType string, keyData []byte) (PublicKey, error) {
	if keyType == "" {
		keyType = "identity"
	}

	key := PublicKey{}
	query := `
		INSERT INTO public_keys (id, user_id, key_data, key_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, user_id::text, key_type, key_data, created_at
	`
	row := s.db.QueryRowContext(ctx, query, uuid.NewString(), userID, keyData, keyType)
	if err := row.Scan(&key.ID, &key.UserID, &key.KeyType, &key.KeyData, &key.CreatedAt); err != nil {
		return PublicKey{}, fmt.Errorf("insert public key: %w", err)
	}
	return key, nil
}

func (s *PostgresKeyStore) GetLatestPublicKey(ctx context.Context, userID, keyType string) (PublicKey, error) {
	if keyType == "" {
		keyType = "identity"
	}

	key := PublicKey{}
	query := `
		SELECT id::text, user_id::text, key_type, key_data, created_at
		FROM public_keys
		WHERE user_id = $1 AND key_type = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := s.db.QueryRowContext(ctx, query, userID, keyType)
	if err := row.Scan(&key.ID, &key.UserID, &key.KeyType, &key.KeyData, &key.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PublicKey{}, ErrKeyNotFound
		}
		return PublicKey{}, fmt.Errorf("query public key: %w", err)
	}
	return key, nil
}

func NewKeyStoreFromEnv() (KeyStore, func() error, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return NewInMemoryKeyStore(), func() error { return nil }, nil
	}

	postgresStore, err := NewPostgresKeyStore(databaseURL)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			return nil, nil, fmt.Errorf("init postgres key store: %s (%s)", pgErr.Message, pgErr.Code)
		}
		return nil, nil, err
	}

	return postgresStore, postgresStore.Close, nil
}
