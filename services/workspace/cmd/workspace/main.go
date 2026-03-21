package main

import (
	"log"

	"securecollab/services/workspace/internal/handlers"
	"securecollab/services/workspace/internal/store"
)

func main() {
	wsStore, closeStore, err := store.NewStoreFromEnv()
	if err != nil {
		log.Fatalf("failed to initialize workspace store: %v", err)
	}
	defer func() {
		if err := closeStore(); err != nil {
			log.Printf("failed to close workspace store: %v", err)
		}
	}()

	router := handlers.NewRouter(wsStore)
	if err := router.Run(":8086"); err != nil {
		log.Fatalf("workspace service exited: %v", err)
	}
}
