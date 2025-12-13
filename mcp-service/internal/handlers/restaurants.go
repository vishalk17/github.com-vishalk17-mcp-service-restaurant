package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	mw "github.com/vishalk17/mcp-service-restaurant/internal/middleware"
)

type Restaurant struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
	CuisineType string `json:"cuisine_type"`
}

type MenuItem struct {
	ID           int     `json:"id"`
	RestaurantID int     `json:"restaurant_id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	Category     string  `json:"category"`
	DietaryType  string  `json:"dietary_type"`
	SpiceLevel   string  `json:"spice_level"`
	Available    bool    `json:"available"`
}

type RestaurantHandler struct {
	db *sql.DB
}

func NewRestaurantHandler(db *sql.DB) *RestaurantHandler {
	return &RestaurantHandler{db: db}
}

// ListRestaurants handles GET /api/restaurants
func (h *RestaurantHandler) ListRestaurants(w http.ResponseWriter, r *http.Request) {
	if mw.IsDebug() { log.Printf("ListRestaurants called from %s", r.RemoteAddr) }
	rows, err := h.db.Query(`
		SELECT id, name, address, phone_number, cuisine_type 
		FROM restaurants 
		ORDER BY name
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	restaurants := []Restaurant{}
	for rows.Next() {
		var r Restaurant
		if err := rows.Scan(&r.ID, &r.Name, &r.Address, &r.PhoneNumber, &r.CuisineType); err != nil {
			continue
		}
		restaurants = append(restaurants, r)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurants)
}

// GetRestaurant handles GET /api/restaurants/{id}
func (h *RestaurantHandler) GetRestaurant(w http.ResponseWriter, r *http.Request) {
	if mw.IsDebug() { log.Printf("GetRestaurant called from %s", r.RemoteAddr) }
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	var restaurant Restaurant
	err = h.db.QueryRow(`
		SELECT id, name, address, phone_number, cuisine_type 
		FROM restaurants 
		WHERE id = $1
	`, id).Scan(&restaurant.ID, &restaurant.Name, &restaurant.Address, &restaurant.PhoneNumber, &restaurant.CuisineType)

	if err == sql.ErrNoRows {
		http.Error(w, "Restaurant not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurant)
}

// GetMenu handles GET /api/restaurants/{id}/menu
func (h *RestaurantHandler) GetMenu(w http.ResponseWriter, r *http.Request) {
	if mw.IsDebug() { log.Printf("GetMenu called from %s", r.RemoteAddr) }
	idStr := r.URL.Query().Get("restaurant_id")
	if idStr == "" {
		http.Error(w, "Missing restaurant_id parameter", http.StatusBadRequest)
		return
	}

	restaurantID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid restaurant_id", http.StatusBadRequest)
		return
	}

	rows, err := h.db.Query(`
		SELECT id, restaurant_id, name, description, price, category, dietary_type, spice_level, available
		FROM menu_items 
		WHERE restaurant_id = $1 AND available = true
		ORDER BY category, name
	`, restaurantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	menuItems := []MenuItem{}
	for rows.Next() {
		var m MenuItem
		if err := rows.Scan(&m.ID, &m.RestaurantID, &m.Name, &m.Description, &m.Price, &m.Category, &m.DietaryType, &m.SpiceLevel, &m.Available); err != nil {
			continue
		}
		menuItems = append(menuItems, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuItems)
}
