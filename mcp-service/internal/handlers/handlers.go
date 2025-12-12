package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/vishalk17/mcp-service-restaurant/internal/models"
	"github.com/vishalk17/mcp-service-restaurant/internal/storage"
)

type Handlers struct {
	DB *storage.DB
}

func NewHandlers(db *storage.DB) *Handlers {
	return &Handlers{DB: db}
}

// Restaurant Handlers
func (h *Handlers) GetAllRestaurants(w http.ResponseWriter, r *http.Request) {
	restaurants, err := h.DB.GetAllRestaurants()
	if err != nil {
		log.Printf("Error getting restaurants: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurants)
}

func (h *Handlers) GetRestaurantByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	restaurant, err := h.DB.GetRestaurantByID(id)
	if err != nil {
		log.Printf("Error getting restaurant %d: %v", id, err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Restaurant not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurant)
}

func (h *Handlers) CreateRestaurant(w http.ResponseWriter, r *http.Request) {
	var restaurant models.Restaurant
	if err := json.NewDecoder(r.Body).Decode(&restaurant); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	restaurant.CreatedAt = time.Now()
	if err := h.DB.CreateRestaurant(&restaurant); err != nil {
		log.Printf("Error creating restaurant: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(restaurant)
}

func (h *Handlers) UpdateRestaurant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	var restaurant models.Restaurant
	if err := json.NewDecoder(r.Body).Decode(&restaurant); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	restaurant.CreatedAt = time.Now()
	if err := h.DB.UpdateRestaurant(id, &restaurant); err != nil {
		log.Printf("Error updating restaurant %d: %v", id, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurant)
}

func (h *Handlers) DeleteRestaurant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	if err := h.DB.DeleteRestaurant(id); err != nil {
		log.Printf("Error deleting restaurant %d: %v", id, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Menu Item Handlers
func (h *Handlers) GetMenuByRestaurantID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	menuItems, err := h.DB.GetMenuByRestaurantID(restaurantID)
	if err != nil {
		log.Printf("Error getting menu for restaurant %d: %v", restaurantID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuItems)
}

func (h *Handlers) AddMenuItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	var menuItem models.MenuItem
	if err := json.NewDecoder(r.Body).Decode(&menuItem); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	menuItem.RestaurantID = restaurantID
	menuItem.CreatedAt = time.Now()
	if err := h.DB.CreateMenuItem(&menuItem); err != nil {
		log.Printf("Error creating menu item: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(menuItem)
}

// Order Handlers
func (h *Handlers) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.DB.GetAllOrders()
	if err != nil {
		log.Printf("Error getting orders: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *Handlers) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.DB.GetOrderByID(id)
	if err != nil {
		log.Printf("Error getting order %d: %v", id, err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *Handlers) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Calculate amounts for billing
	order.TotalAmount = 0
	for i := range order.OrderItems {
		// Calculate subtotal for each item
		order.OrderItems[i].Subtotal = float64(order.OrderItems[i].Quantity) * order.OrderItems[i].Price
		order.TotalAmount += order.OrderItems[i].Subtotal
	}

	// Apply 5% GST tax (simplified for demo)
	order.TaxAmount = order.TotalAmount * 0.05
	order.FinalAmount = order.TotalAmount + order.TaxAmount - order.Discount

	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	if err := h.DB.CreateOrder(&order); err != nil {
		log.Printf("Error creating order: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Fetch the complete order with all details after creation
	completeOrder, err := h.DB.GetOrderByID(order.ID)
	if err != nil {
		log.Printf("Error fetching created order: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(completeOrder)
}

// Health check handler
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}