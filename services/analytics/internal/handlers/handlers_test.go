package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"securecollab/services/analytics/internal/store"
)

type fakeStore struct {
	volume store.MessageVolume
	err    error
}

func (f *fakeStore) GetMessageVolume(_ context.Context, windowHours int) (store.MessageVolume, error) {
	if f.err != nil {
		return store.MessageVolume{}, f.err
	}
	v := f.volume
	if v.WindowHours == 0 {
		v.WindowHours = windowHours
	}
	return v, nil
}

func TestMessageVolume_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(&fakeStore{volume: store.MessageVolume{TotalMessages: 10, MessagesLast24h: 4, WindowHours: 24}})
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/v1/analytics/messages/volume", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestMessageVolume_InvalidWindow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(&fakeStore{})
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/v1/analytics/messages/volume?window_hours=0", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestMessageVolume_StoreError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(&fakeStore{err: errors.New("boom")})
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/v1/analytics/messages/volume", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.Code)
	}
}
