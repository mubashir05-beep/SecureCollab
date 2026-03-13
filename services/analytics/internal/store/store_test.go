package store

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryAnalyticsStore_GetMessageVolume(t *testing.T) {
	s := NewInMemoryAnalyticsStore()
	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	s.clockFn = func() time.Time { return now }
	s.SetMessageTimes([]time.Time{
		now.Add(-1 * time.Hour),
		now.Add(-2 * time.Hour),
		now.Add(-30 * time.Hour),
	})

	volume, err := s.GetMessageVolume(context.Background(), 24)
	if err != nil {
		t.Fatalf("get volume: %v", err)
	}
	if volume.TotalMessages != 3 {
		t.Fatalf("expected total 3, got %d", volume.TotalMessages)
	}
	if volume.MessagesLast24h != 2 {
		t.Fatalf("expected 2 messages in 24h, got %d", volume.MessagesLast24h)
	}
}
