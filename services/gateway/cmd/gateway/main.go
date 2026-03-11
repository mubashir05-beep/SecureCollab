package main

import (
	"log"

	"securecollab/services/gateway/internal/httpserver"
)

func main() {
	router := httpserver.NewRouter()
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("gateway exited: %v", err)
	}
}
