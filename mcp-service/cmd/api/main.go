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

	// Restaurant routes (at root level, gateway will add /api prefix)
	router.HandleFunc("/restaurants", handlers.GetAllRestaurants).Methods("GET")
	router.HandleFunc("/restaurants", handlers.CreateRestaurant).Methods("POST")
	router.HandleFunc("/restaurants/{id}", handlers.GetRestaurantByID).Methods("GET")
	router.HandleFunc("/restaurants/{id}", handlers.UpdateRestaurant).Methods("PUT")
	router.HandleFunc("/restaurants/{id}", handlers.DeleteRestaurant).Methods("DELETE")

	// Menu routes - nested under restaurant
	router.HandleFunc("/restaurants/{id}/menu", handlers.GetMenuByRestaurantID).Methods("GET")
	router.HandleFunc("/restaurants/{id}/menu", handlers.AddMenuItem).Methods("POST")

	// Order routes
	router.HandleFunc("/orders", handlers.GetAllOrders).Methods("GET")
	router.HandleFunc("/orders", handlers.CreateOrder).Methods("POST")
	router.HandleFunc("/orders/{id}", handlers.GetOrderByID).Methods("GET")

	// Create a custom ServeMux to handle the MCP WebSocket endpoint at the server level
	mainMux := http.NewServeMux()

	// Handle MCP WebSocket endpoint with custom handler to ensure proper upgrade
	// This is registered BEFORE the router, so it takes priority
	mainMux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a WebSocket upgrade request
		if r.Header.Get("Connection") == "Upgrade" && r.Header.Get("Upgrade") == "websocket" {
			// This will be handled by the existing MCPWebSocketHandler
			handlers.MCPWebSocketHandler(w, r)
		} else {
			// For any other request to /mcp, return a method not allowed error
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Handle MCP with trailing slash as well
	mainMux.HandleFunc("/mcp/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Connection") == "Upgrade" && r.Header.Get("Upgrade") == "websocket" {
			handlers.MCPWebSocketHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Add health check at root
	mainMux.Handle("/health", http.HandlerFunc(handlers.HealthCheck))

	// Add all other routes through router (no prefix stripping needed)
	mainMux.Handle("/", router)

	// Server configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("MCP Service starting on port %s\n", port)
	fmt.Printf("Database connected successfully\n")

	// Start server with the custom multiplexer
	log.Fatal(http.ListenAndServe(":"+port, mainMux))
}