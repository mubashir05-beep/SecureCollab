package main

import (
	"log"

	"securecollab/services/messaging/internal/handlers"
	"securecollab/services/messaging/internal/store"
)

func main() {
	messageStore, closeStore, err := store.NewMessageStoreFromEnv()
	if err != nil {
		log.Fatalf("failed to initialize message store: %v", err)
	}
	defer func() {
		if err := closeStore(); err != nil {
			log.Printf("failed to close message store: %v", err)
		}
	}()

	router := handlers.NewRouter(messageStore)
	if err := router.Run(":8083"); err != nil {
		log.Fatalf("messaging service exited: %v", err)
	}
}
