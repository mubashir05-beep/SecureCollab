package main

import (
	"log"

	"securecollab/services/auth/internal/handlers"
	"securecollab/services/auth/internal/store"
)

func main() {
	userStore, closeStore, err := store.NewUserStoreFromEnv()
	if err != nil {
		log.Fatalf("failed to initialize user store: %v", err)
	}
	defer func() {
		if err := closeStore(); err != nil {
			log.Printf("failed to close user store: %v", err)
		}
	}()

	router := handlers.NewRouter(userStore)
	if err := router.Run(":8081"); err != nil {
		log.Fatalf("auth service exited: %v", err)
	}
}
