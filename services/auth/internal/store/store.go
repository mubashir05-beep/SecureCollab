package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var ErrUserExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
}

type UserStore interface {
	CreateUser(ctx context.Context, username, email, passwordHash string) (User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
}

type InMemoryUserStore struct {
	byUsername map[string]User
	byEmail    map[string]User
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		byUsername: make(map[string]User),
		byEmail:    make(map[string]User),
	}
}

func (s *InMemoryUserStore) CreateUser(_ context.Context, username, email, passwordHash string) (User, error) {
	if _, ok := s.byUsername[username]; ok {
		return User{}, ErrUserExists
	}
	if _, ok := s.byEmail[email]; ok {
		return User{}, ErrUserExists
	}

	user := User{
		ID:           uuid.NewString(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}
	s.byUsername[username] = user
	s.byEmail[email] = user
	return user, nil
}

func (s *InMemoryUserStore) GetUserByUsername(_ context.Context, username string) (User, error) {
	user, ok := s.byUsername[username]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

type PostgresUserStore struct {
	db *sql.DB
}

func NewPostgresUserStore(databaseURL string) (*PostgresUserStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &PostgresUserStore{db: db}, nil
}

func (s *PostgresUserStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *PostgresUserStore) CreateUser(ctx context.Context, username, email, passwordHash string) (User, error) {
	user := User{}
	query := `
		INSERT INTO users (id, username, email, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, username, email, password_hash
	`
	row := s.db.QueryRowContext(ctx, query, uuid.NewString(), username, email, passwordHash)
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash); err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, ErrUserExists
		}
		return User{}, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

func (s *PostgresUserStore) GetUserByUsername(ctx context.Context, username string) (User, error) {
	user := User{}
	query := `
		SELECT id::text, username, email, password_hash
		FROM users
		WHERE username = $1
	`
	row := s.db.QueryRowContext(ctx, query, username)
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("query user: %w", err)
	}
	return user, nil
}

func NewUserStoreFromEnv() (UserStore, func() error, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return NewInMemoryUserStore(), func() error { return nil }, nil
	}

	postgresStore, err := NewPostgresUserStore(databaseURL)
	if err != nil {
		return nil, nil, err
	}

	return postgresStore, postgresStore.Close, nil
}
