package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"securecollab/services/analytics/internal/handlers"
	"securecollab/services/analytics/internal/store"
)

func main() {
	analyticsStore, closeStore, err := store.NewAnalyticsStoreWithClickHouse()
	if err != nil {
		log.Fatalf("failed to initialize analytics store: %v", err)
	}
	defer func() {
		if err := closeStore(); err != nil {
			log.Printf("failed to close analytics store: %v", err)
		}
	}()

	router := gin.Default()
	h := handlers.NewHandler(analyticsStore)
	h.RegisterRoutes(router)

	if err := router.Run(":8084"); err != nil {
		log.Fatalf("analytics service exited: %v", err)
	}
}
