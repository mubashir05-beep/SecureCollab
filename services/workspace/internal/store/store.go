package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrChannelNotFound   = errors.New("channel not found")
	ErrAlreadyMember     = errors.New("already a member")
	ErrNotMember         = errors.New("not a member")
	ErrForbidden         = errors.New("forbidden")
	ErrChannelExists     = errors.New("channel already exists")
	ErrInvalidInvite     = errors.New("invalid invite code")
)

type Workspace struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	InviteCode  string    `json:"invite_code,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type WorkspaceMember struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

var ErrUserNotFound = errors.New("user not found")

type Channel struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Topic       string     `json:"topic"`
	IsPrivate   bool       `json:"is_private"`
	CreatedBy   string     `json:"created_by"`
	ArchivedAt  *time.Time `json:"archived_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type WorkspaceStore interface {
	CreateWorkspace(ctx context.Context, name, description, ownerID string) (Workspace, error)
	GetWorkspace(ctx context.Context, id string) (Workspace, error)
	ListWorkspacesForUser(ctx context.Context, userID string) ([]Workspace, error)
	JoinWorkspaceByInvite(ctx context.Context, inviteCode, userID string) (Workspace, error)
	AddMember(ctx context.Context, workspaceID, userID, role string) error
	RemoveMember(ctx context.Context, workspaceID, userID string) error
	ListMembers(ctx context.Context, workspaceID string) ([]WorkspaceMember, error)
	GetMemberRole(ctx context.Context, workspaceID, userID string) (string, error)
	ResolveUserID(ctx context.Context, usernameOrID string) (string, error)

	CreateChannel(ctx context.Context, workspaceID, name, description, topic, createdBy string, isPrivate bool) (Channel, error)
	GetChannel(ctx context.Context, id string) (Channel, error)
	ListChannels(ctx context.Context, workspaceID string) ([]Channel, error)
	UpdateChannelTopic(ctx context.Context, channelID, topic string) error
	ArchiveChannel(ctx context.Context, channelID string) error
}

// --- In-Memory Implementation ---

type inMemoryStore struct {
	mu         sync.Mutex
	workspaces map[string]Workspace
	members    map[string][]WorkspaceMember // workspaceID -> members
	channels   map[string]Channel
	chanByWS   map[string][]string // workspaceID -> channelIDs
}

func NewInMemoryStore() WorkspaceStore {
	return &inMemoryStore{
		workspaces: make(map[string]Workspace),
		members:    make(map[string][]WorkspaceMember),
		channels:   make(map[string]Channel),
		chanByWS:   make(map[string][]string),
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateInviteCode() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *inMemoryStore) CreateWorkspace(_ context.Context, name, description, ownerID string) (Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ws := Workspace{
		ID:          generateID(),
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		InviteCode:  generateInviteCode(),
		CreatedAt:   time.Now(),
	}
	s.workspaces[ws.ID] = ws
	s.members[ws.ID] = []WorkspaceMember{{UserID: ownerID, Role: "owner", JoinedAt: time.Now()}}
	return ws, nil
}

func (s *inMemoryStore) GetWorkspace(_ context.Context, id string) (Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ws, ok := s.workspaces[id]
	if !ok {
		return Workspace{}, ErrWorkspaceNotFound
	}
	return ws, nil
}

func (s *inMemoryStore) ListWorkspacesForUser(_ context.Context, userID string) ([]Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []Workspace
	for wsID, members := range s.members {
		for _, m := range members {
			if m.UserID == userID {
				if ws, ok := s.workspaces[wsID]; ok {
					result = append(result, ws)
				}
				break
			}
		}
	}
	return result, nil
}

func (s *inMemoryStore) JoinWorkspaceByInvite(_ context.Context, inviteCode, userID string) (Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ws := range s.workspaces {
		if ws.InviteCode == inviteCode {
			for _, m := range s.members[ws.ID] {
				if m.UserID == userID {
					return ws, ErrAlreadyMember
				}
			}
			s.members[ws.ID] = append(s.members[ws.ID], WorkspaceMember{UserID: userID, Role: "member", JoinedAt: time.Now()})
			return ws, nil
		}
	}
	return Workspace{}, ErrInvalidInvite
}

func (s *inMemoryStore) AddMember(_ context.Context, workspaceID, userID, role string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, m := range s.members[workspaceID] {
		if m.UserID == userID {
			return ErrAlreadyMember
		}
	}
	s.members[workspaceID] = append(s.members[workspaceID], WorkspaceMember{UserID: userID, Role: role, JoinedAt: time.Now()})
	return nil
}

func (s *inMemoryStore) RemoveMember(_ context.Context, workspaceID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	members := s.members[workspaceID]
	for i, m := range members {
		if m.UserID == userID {
			s.members[workspaceID] = append(members[:i], members[i+1:]...)
			return nil
		}
	}
	return ErrNotMember
}

func (s *inMemoryStore) ListMembers(_ context.Context, workspaceID string) ([]WorkspaceMember, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.members[workspaceID], nil
}

func (s *inMemoryStore) GetMemberRole(_ context.Context, workspaceID, userID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, m := range s.members[workspaceID] {
		if m.UserID == userID {
			return m.Role, nil
		}
	}
	return "", ErrNotMember
}

func (s *inMemoryStore) ResolveUserID(_ context.Context, usernameOrID string) (string, error) {
	return usernameOrID, nil
}

func (s *inMemoryStore) CreateChannel(_ context.Context, workspaceID, name, description, topic, createdBy string, isPrivate bool) (Channel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, chID := range s.chanByWS[workspaceID] {
		if ch, ok := s.channels[chID]; ok && ch.Name == name {
			return Channel{}, ErrChannelExists
		}
	}
	ch := Channel{
		ID:          generateID(),
		WorkspaceID: workspaceID,
		Name:        name,
		Description: description,
		Topic:       topic,
		IsPrivate:   isPrivate,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
	}
	s.channels[ch.ID] = ch
	s.chanByWS[workspaceID] = append(s.chanByWS[workspaceID], ch.ID)
	return ch, nil
}

func (s *inMemoryStore) GetChannel(_ context.Context, id string) (Channel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch, ok := s.channels[id]
	if !ok {
		return Channel{}, ErrChannelNotFound
	}
	return ch, nil
}

func (s *inMemoryStore) ListChannels(_ context.Context, workspaceID string) ([]Channel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []Channel
	for _, chID := range s.chanByWS[workspaceID] {
		if ch, ok := s.channels[chID]; ok && ch.ArchivedAt == nil {
			result = append(result, ch)
		}
	}
	return result, nil
}

func (s *inMemoryStore) UpdateChannelTopic(_ context.Context, channelID, topic string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch, ok := s.channels[channelID]
	if !ok {
		return ErrChannelNotFound
	}
	ch.Topic = topic
	s.channels[channelID] = ch
	return nil
}

func (s *inMemoryStore) ArchiveChannel(_ context.Context, channelID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch, ok := s.channels[channelID]
	if !ok {
		return ErrChannelNotFound
	}
	now := time.Now()
	ch.ArchivedAt = &now
	s.channels[channelID] = ch
	return nil
}

// --- Postgres Implementation ---

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) CreateWorkspace(ctx context.Context, name, description, ownerID string) (Workspace, error) {
	inviteCode := generateInviteCode()
	var ws Workspace
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO workspaces (name, description, owner_id, invite_code) VALUES ($1, $2, $3, $4)
		 RETURNING id, name, description, owner_id, invite_code, created_at`,
		name, description, ownerID, inviteCode,
	).Scan(&ws.ID, &ws.Name, &ws.Description, &ws.OwnerID, &ws.InviteCode, &ws.CreatedAt)
	if err != nil {
		return Workspace{}, fmt.Errorf("insert workspace: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'owner')`,
		ws.ID, ownerID)
	if err != nil {
		return Workspace{}, fmt.Errorf("insert owner member: %w", err)
	}
	return ws, nil
}

func (s *PostgresStore) GetWorkspace(ctx context.Context, id string) (Workspace, error) {
	var ws Workspace
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, COALESCE(description,''), owner_id, COALESCE(invite_code,''), created_at FROM workspaces WHERE id = $1`, id,
	).Scan(&ws.ID, &ws.Name, &ws.Description, &ws.OwnerID, &ws.InviteCode, &ws.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Workspace{}, ErrWorkspaceNotFound
		}
		return Workspace{}, err
	}
	return ws, nil
}

func (s *PostgresStore) ListWorkspacesForUser(ctx context.Context, userID string) ([]Workspace, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT w.id, w.name, COALESCE(w.description,''), w.owner_id, COALESCE(w.invite_code,''), w.created_at
		 FROM workspaces w JOIN workspace_members wm ON w.id = wm.workspace_id
		 WHERE wm.user_id = $1 ORDER BY w.created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Workspace
	for rows.Next() {
		var ws Workspace
		if err := rows.Scan(&ws.ID, &ws.Name, &ws.Description, &ws.OwnerID, &ws.InviteCode, &ws.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, ws)
	}
	return result, rows.Err()
}

func (s *PostgresStore) JoinWorkspaceByInvite(ctx context.Context, inviteCode, userID string) (Workspace, error) {
	var ws Workspace
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, COALESCE(description,''), owner_id, invite_code, created_at FROM workspaces WHERE invite_code = $1`, inviteCode,
	).Scan(&ws.ID, &ws.Name, &ws.Description, &ws.OwnerID, &ws.InviteCode, &ws.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Workspace{}, ErrInvalidInvite
		}
		return Workspace{}, err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'member') ON CONFLICT DO NOTHING`, ws.ID, userID)
	if err != nil {
		return Workspace{}, err
	}
	return ws, nil
}

func (s *PostgresStore) AddMember(ctx context.Context, workspaceID, userID, role string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, $3)`, workspaceID, userID, role)
	if err != nil && strings.Contains(err.Error(), "duplicate") {
		return ErrAlreadyMember
	}
	return err
}

func (s *PostgresStore) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`, workspaceID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotMember
	}
	return nil
}

func (s *PostgresStore) ListMembers(ctx context.Context, workspaceID string) ([]WorkspaceMember, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT wm.user_id, COALESCE(u.username, ''), wm.role, wm.joined_at
		 FROM workspace_members wm
		 LEFT JOIN users u ON u.id = wm.user_id
		 WHERE wm.workspace_id = $1 ORDER BY wm.joined_at`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []WorkspaceMember
	for rows.Next() {
		var m WorkspaceMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (s *PostgresStore) GetMemberRole(ctx context.Context, workspaceID, userID string) (string, error) {
	var role string
	err := s.db.QueryRowContext(ctx,
		`SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`, workspaceID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotMember
		}
		return "", err
	}
	return role, nil
}

func (s *PostgresStore) ResolveUserID(ctx context.Context, usernameOrID string) (string, error) {
	// If it looks like a UUID, verify it exists and return it.
	if isUUID(usernameOrID) {
		var id string
		err := s.db.QueryRowContext(ctx, `SELECT id::text FROM users WHERE id = $1`, usernameOrID).Scan(&id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", ErrUserNotFound
			}
			return "", err
		}
		return id, nil
	}
	// Otherwise treat as username.
	var id string
	err := s.db.QueryRowContext(ctx, `SELECT id::text FROM users WHERE username = $1`, usernameOrID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}
	return id, nil
}

func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func (s *PostgresStore) CreateChannel(ctx context.Context, workspaceID, name, description, topic, createdBy string, isPrivate bool) (Channel, error) {
	var ch Channel
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO channels (workspace_id, name, description, topic, is_private, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, workspace_id, name, COALESCE(description,''), COALESCE(topic,''), is_private, created_by, created_at`,
		workspaceID, name, description, topic, isPrivate, createdBy,
	).Scan(&ch.ID, &ch.WorkspaceID, &ch.Name, &ch.Description, &ch.Topic, &ch.IsPrivate, &ch.CreatedBy, &ch.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return Channel{}, ErrChannelExists
		}
		return Channel{}, fmt.Errorf("insert channel: %w", err)
	}
	return ch, nil
}

func (s *PostgresStore) GetChannel(ctx context.Context, id string) (Channel, error) {
	var ch Channel
	err := s.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, name, COALESCE(description,''), COALESCE(topic,''), is_private, COALESCE(created_by::text,''), archived_at, created_at
		 FROM channels WHERE id = $1`, id,
	).Scan(&ch.ID, &ch.WorkspaceID, &ch.Name, &ch.Description, &ch.Topic, &ch.IsPrivate, &ch.CreatedBy, &ch.ArchivedAt, &ch.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Channel{}, ErrChannelNotFound
		}
		return Channel{}, err
	}
	return ch, nil
}

func (s *PostgresStore) ListChannels(ctx context.Context, workspaceID string) ([]Channel, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, workspace_id, name, COALESCE(description,''), COALESCE(topic,''), is_private, COALESCE(created_by::text,''), created_at
		 FROM channels WHERE workspace_id = $1 AND archived_at IS NULL ORDER BY created_at`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Channel
	for rows.Next() {
		var ch Channel
		if err := rows.Scan(&ch.ID, &ch.WorkspaceID, &ch.Name, &ch.Description, &ch.Topic, &ch.IsPrivate, &ch.CreatedBy, &ch.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, ch)
	}
	return result, rows.Err()
}

func (s *PostgresStore) UpdateChannelTopic(ctx context.Context, channelID, topic string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE channels SET topic = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, topic, channelID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrChannelNotFound
	}
	return nil
}

func (s *PostgresStore) ArchiveChannel(ctx context.Context, channelID string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE channels SET archived_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, channelID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrChannelNotFound
	}
	return nil
}

func NewStoreFromEnv() (WorkspaceStore, func() error, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return NewInMemoryStore(), func() error { return nil }, nil
	}
	pg, err := NewPostgresStore(databaseURL)
	if err != nil {
		return nil, nil, err
	}
	return pg, pg.Close, nil
}
