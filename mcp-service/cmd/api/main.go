package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/vishalk17/mcp-service-restaurant/internal/handlers"
	"github.com/vishalk17/mcp-service-restaurant/internal/storage"
)

func main() {
	// Get database connection string from environment variable
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Fallback to local development connection string
		dbURL = "host=localhost port=5432 user=postgres password=postgres dbname=mcp_restaurant sslmode=disable"
	}

	// Initialize database
	db, err := storage.NewDB(dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize handlers
	handlers := handlers.NewHandlers(db)

	// Create router
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

	// Main MCP endpoint group
	api := router.PathPrefix("/mcp").Subrouter()

	// Restaurant routes
	api.HandleFunc("/restaurants", handlers.GetAllRestaurants).Methods("GET")
	api.HandleFunc("/restaurants", handlers.CreateRestaurant).Methods("POST")
	api.HandleFunc("/restaurants/{id}", handlers.GetRestaurantByID).Methods("GET")
	api.HandleFunc("/restaurants/{id}", handlers.UpdateRestaurant).Methods("PUT")
	api.HandleFunc("/restaurants/{id}", handlers.DeleteRestaurant).Methods("DELETE")

	// Menu routes - nested under restaurant
	api.HandleFunc("/restaurants/{id}/menu", handlers.GetMenuByRestaurantID).Methods("GET")
	api.HandleFunc("/restaurants/{id}/menu", handlers.AddMenuItem).Methods("POST")

	// Order routes
	api.HandleFunc("/orders", handlers.GetAllOrders).Methods("GET")
	api.HandleFunc("/orders", handlers.CreateOrder).Methods("POST")
	api.HandleFunc("/orders/{id}", handlers.GetOrderByID).Methods("GET")

	// Server configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("MCP Service starting on port %s\n", port)
	fmt.Printf("Database connected successfully\n")
	
	// Start server
	log.Fatal(http.ListenAndServe(":"+port, router))
}