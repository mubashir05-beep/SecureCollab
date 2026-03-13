package main

import (
	"log"

	"securecollab/services/keydist/internal/handlers"
	"securecollab/services/keydist/internal/store"
)

func main() {
	keyStore, closeStore, err := store.NewKeyStoreFromEnv()
	if err != nil {
		log.Fatalf("failed to initialize key store: %v", err)
	}
	defer func() {
		if err := closeStore(); err != nil {
			log.Printf("failed to close key store: %v", err)
		}
	}()

	router := handlers.NewRouter(keyStore)
	if err := router.Run(":8082"); err != nil {
		log.Fatalf("key distribution service exited: %v", err)
	}
}
