package main

import (
	"log"

	"securecollab/services/auth/internal/handlers"
)

func main() {
	router := handlers.NewRouter()
	if err := router.Run(":8081"); err != nil {
		log.Fatalf("auth service exited: %v", err)
	}
}
