package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"securecollab/services/workspace/internal/store"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "securecollab-dev-secret-key"

func generateTestToken(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	s, _ := token.SignedString([]byte(testSecret))
	return s
}

func setupRouter() *testEnv {
	s := store.NewInMemoryStore()
	r := NewRouter(s)
	return &testEnv{store: s, router: r}
}

type testEnv struct {
	store  store.WorkspaceStore
	router http.Handler
}

func (te *testEnv) request(method, path, token string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	te.router.ServeHTTP(w, req)
	return w
}

func TestHealthz(t *testing.T) {
	te := setupRouter()
	w := te.request("GET", "/healthz", "", nil)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCreateWorkspace(t *testing.T) {
	te := setupRouter()
	token := generateTestToken("user-1")

	w := te.request("POST", "/v1/workspaces", token, CreateWorkspaceRequest{
		Name:        "Test Workspace",
		Description: "A test workspace",
	})
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var ws store.Workspace
	json.Unmarshal(w.Body.Bytes(), &ws)
	if ws.Name != "Test Workspace" {
		t.Fatalf("expected name 'Test Workspace', got '%s'", ws.Name)
	}
	if ws.OwnerID != "user-1" {
		t.Fatalf("expected owner_id 'user-1', got '%s'", ws.OwnerID)
	}
	if ws.InviteCode == "" {
		t.Fatal("expected invite_code to be set")
	}
}

func TestListWorkspaces(t *testing.T) {
	te := setupRouter()
	token := generateTestToken("user-1")

	// Create two workspaces
	te.request("POST", "/v1/workspaces", token, CreateWorkspaceRequest{Name: "WS1"})
	te.request("POST", "/v1/workspaces", token, CreateWorkspaceRequest{Name: "WS2"})

	w := te.request("GET", "/v1/workspaces", token, nil)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Workspaces []store.Workspace `json:"workspaces"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Workspaces) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(resp.Workspaces))
	}
}

func TestJoinByInvite(t *testing.T) {
	te := setupRouter()
	ownerToken := generateTestToken("owner-1")
	memberToken := generateTestToken("member-1")

	// Owner creates workspace
	w := te.request("POST", "/v1/workspaces", ownerToken, CreateWorkspaceRequest{Name: "Invite WS"})
	var ws store.Workspace
	json.Unmarshal(w.Body.Bytes(), &ws)

	// Member joins by invite
	w = te.request("POST", "/v1/workspaces/join", memberToken, JoinByInviteRequest{InviteCode: ws.InviteCode})
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Member should see it in their list
	w = te.request("GET", "/v1/workspaces", memberToken, nil)
	var resp struct {
		Workspaces []store.Workspace `json:"workspaces"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(resp.Workspaces))
	}
}

func TestCreateAndListChannels(t *testing.T) {
	te := setupRouter()
	token := generateTestToken("user-1")

	// Create workspace
	w := te.request("POST", "/v1/workspaces", token, CreateWorkspaceRequest{Name: "Chan WS"})
	var ws store.Workspace
	json.Unmarshal(w.Body.Bytes(), &ws)

	// Create channels
	te.request("POST", fmt.Sprintf("/v1/workspaces/%s/channels", ws.ID), token,
		CreateChannelRequest{Name: "general", Topic: "General discussion"})
	te.request("POST", fmt.Sprintf("/v1/workspaces/%s/channels", ws.ID), token,
		CreateChannelRequest{Name: "secret", IsPrivate: true})

	// List channels
	w = te.request("GET", fmt.Sprintf("/v1/workspaces/%s/channels", ws.ID), token, nil)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Channels []store.Channel `json:"channels"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(resp.Channels))
	}
}

func TestMemberManagement(t *testing.T) {
	te := setupRouter()
	ownerToken := generateTestToken("owner-1")

	// Create workspace
	w := te.request("POST", "/v1/workspaces", ownerToken, CreateWorkspaceRequest{Name: "Members WS"})
	var ws store.Workspace
	json.Unmarshal(w.Body.Bytes(), &ws)

	// Add member
	w = te.request("POST", fmt.Sprintf("/v1/workspaces/%s/members", ws.ID), ownerToken,
		AddMemberRequest{UserID: "member-1", Role: "member"})
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// List members
	w = te.request("GET", fmt.Sprintf("/v1/workspaces/%s/members", ws.ID), ownerToken, nil)
	var resp struct {
		Members []store.WorkspaceMember `json:"members"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Members) != 2 { // owner + member
		t.Fatalf("expected 2 members, got %d", len(resp.Members))
	}

	// Remove member
	w = te.request("DELETE", fmt.Sprintf("/v1/workspaces/%s/members/member-1", ws.ID), ownerToken, nil)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestNonMemberCannotAddMembers(t *testing.T) {
	te := setupRouter()
	ownerToken := generateTestToken("owner-1")
	randomToken := generateTestToken("random-user")

	w := te.request("POST", "/v1/workspaces", ownerToken, CreateWorkspaceRequest{Name: "Protected WS"})
	var ws store.Workspace
	json.Unmarshal(w.Body.Bytes(), &ws)

	// Random user tries to add member -> forbidden
	w = te.request("POST", fmt.Sprintf("/v1/workspaces/%s/members", ws.ID), randomToken,
		AddMemberRequest{UserID: "attacker"})
	if w.Code != 403 {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestArchiveChannel(t *testing.T) {
	te := setupRouter()
	token := generateTestToken("user-1")

	w := te.request("POST", "/v1/workspaces", token, CreateWorkspaceRequest{Name: "Archive WS"})
	var ws store.Workspace
	json.Unmarshal(w.Body.Bytes(), &ws)

	w = te.request("POST", fmt.Sprintf("/v1/workspaces/%s/channels", ws.ID), token,
		CreateChannelRequest{Name: "temp-channel"})
	var ch store.Channel
	json.Unmarshal(w.Body.Bytes(), &ch)

	// Archive
	w = te.request("POST", fmt.Sprintf("/v1/channels/%s/archive", ch.ID), token, nil)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Should not appear in list
	w = te.request("GET", fmt.Sprintf("/v1/workspaces/%s/channels", ws.ID), token, nil)
	var resp struct {
		Channels []store.Channel `json:"channels"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Channels) != 0 {
		t.Fatalf("expected 0 channels after archive, got %d", len(resp.Channels))
	}
}

func TestUnauthorizedRequests(t *testing.T) {
	te := setupRouter()

	w := te.request("GET", "/v1/workspaces", "", nil)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	w = te.request("POST", "/v1/workspaces", "", nil)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
