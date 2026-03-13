package handlers

import "sync"

type deliveryHub struct {
	mu      sync.RWMutex
	clients map[string]map[chan MessageEnvelope]struct{}
}

func newDeliveryHub() *deliveryHub {
	return &deliveryHub{clients: make(map[string]map[chan MessageEnvelope]struct{})}
}

func (h *deliveryHub) Subscribe(userID string) chan MessageEnvelope {
	ch := make(chan MessageEnvelope, 32)

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[userID]; !ok {
		h.clients[userID] = make(map[chan MessageEnvelope]struct{})
	}
	h.clients[userID][ch] = struct{}{}
	return ch
}

func (h *deliveryHub) Unsubscribe(userID string, ch chan MessageEnvelope) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if entries, ok := h.clients[userID]; ok {
		if _, exists := entries[ch]; exists {
			delete(entries, ch)
			close(ch)
		}
		if len(entries) == 0 {
			delete(h.clients, userID)
		}
	}
}

func (h *deliveryHub) Publish(userID string, msg MessageEnvelope) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.clients[userID] {
		select {
		case ch <- msg:
		default:
			// Drop when consumer is too slow to avoid blocking producers.
		}
	}
}
