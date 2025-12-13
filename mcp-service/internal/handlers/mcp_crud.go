package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
)

// Restaurant CRUD
func (h *MCPHandler) toolCreateRestaurant(id interface{}, args map[string]interface{}) MCPResponse {
	name, _ := args["name"].(string)
	address, _ := args["address"].(string)
	phone, _ := args["phone_number"].(string)
	cuisine, _ := args["cuisine_type"].(string)
	
	if cuisine == "" {
		cuisine = "Indian"
	}
	
	var newID int
	err := h.db.QueryRow(`
		INSERT INTO restaurants (name, address, phone_number, cuisine_type)
		VALUES ($1, $2, $3, $4) RETURNING id
	`, name, address, phone, cuisine).Scan(&newID)
	
	if err != nil {
		log.Printf("Error creating restaurant: %v", err)
		return h.errorResponse(id, -32603, "Database error")
	}
	
	return h.successResponse(id, fmt.Sprintf("Restaurant created with ID %d", newID))
}

func (h *MCPHandler) toolUpdateRestaurant(id interface{}, args map[string]interface{}) MCPResponse {
	restaurantID, ok := args["id"].(float64)
	if !ok {
		return h.errorResponse(id, -32602, "Missing id")
	}
	
	name, _ := args["name"].(string)
	address, _ := args["address"].(string)
	phone, _ := args["phone_number"].(string)
	cuisine, _ := args["cuisine_type"].(string)
	
	_, err := h.db.Exec(`
		UPDATE restaurants 
		SET name = COALESCE(NULLIF($1, ''), name),
		    address = COALESCE(NULLIF($2, ''), address),
		    phone_number = COALESCE(NULLIF($3, ''), phone_number),
		    cuisine_type = COALESCE(NULLIF($4, ''), cuisine_type)
		WHERE id = $5
	`, name, address, phone, cuisine, int(restaurantID))
	
	if err != nil {
		log.Printf("Error updating restaurant: %v", err)
		return h.errorResponse(id, -32603, "Database error")
	}
	
	return h.successResponse(id, fmt.Sprintf("Restaurant %d updated", int(restaurantID)))
}

func (h *MCPHandler) toolDeleteRestaurant(id interface{}, args map[string]interface{}) MCPResponse {
	restaurantID, ok := args["id"].(float64)
	if !ok {
		return h.errorResponse(id, -32602, "Missing id")
	}
	
	_, err := h.db.Exec("DELETE FROM restaurants WHERE id = $1", int(restaurantID))
	if err != nil {
		log.Printf("Error deleting restaurant: %v", err)
		return h.errorResponse(id, -32603, "Database error")
	}
	
	return h.successResponse(id, fmt.Sprintf("Restaurant %d deleted", int(restaurantID)))
}

// Menu Item CRUD
func (h *MCPHandler) toolCreateMenuItem(id interface{}, args map[string]interface{}) MCPResponse {
	restaurantID, _ := args["restaurant_id"].(float64)
	name, _ := args["name"].(string)
	description, _ := args["description"].(string)
	price, _ := args["price"].(float64)
	category, _ := args["category"].(string)
	dietary, _ := args["dietary_type"].(string)
	spice, _ := args["spice_level"].(string)
	
	if category == "" {
		category = "Main Course"
	}
	if dietary == "" {
		dietary = "vegetarian"
	}
	if spice == "" {
		spice = "medium"
	}
	
	var newID int
	err := h.db.QueryRow(`
		INSERT INTO menu_items (restaurant_id, name, description, price, category, dietary_type, spice_level, available)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true) RETURNING id
	`, int(restaurantID), name, description, price, category, dietary, spice).Scan(&newID)
	
	if err != nil {
		log.Printf("Error creating menu item: %v", err)
		return h.errorResponse(id, -32603, "Database error")
	}
	
	return h.successResponse(id, fmt.Sprintf("Menu item created with ID %d", newID))
}

func (h *MCPHandler) toolUpdateMenuItem(id interface{}, args map[string]interface{}) MCPResponse {
	menuItemID, ok := args["id"].(float64)
	if !ok {
		return h.errorResponse(id, -32602, "Missing id")
	}
	
	name, _ := args["name"].(string)
	description, _ := args["description"].(string)
	price, _ := args["price"].(float64)
	category, _ := args["category"].(string)
	
	_, err := h.db.Exec(`
		UPDATE menu_items 
		SET name = COALESCE(NULLIF($1, ''), name),
		    description = COALESCE(NULLIF($2, ''), description),
		    price = CASE WHEN $3 > 0 THEN $3 ELSE price END,
		    category = COALESCE(NULLIF($4, ''), category)
		WHERE id = $5
	`, name, description, price, category, int(menuItemID))
	
	if err != nil {
		log.Printf("Error updating menu item: %v", err)
		return h.errorResponse(id, -32603, "Database error")
	}
	
	return h.successResponse(id, fmt.Sprintf("Menu item %d updated", int(menuItemID)))
}

func (h *MCPHandler) toolDeleteMenuItem(id interface{}, args map[string]interface{}) MCPResponse {
	menuItemID, ok := args["id"].(float64)
	if !ok {
		return h.errorResponse(id, -32602, "Missing id")
	}
	
	_, err := h.db.Exec("DELETE FROM menu_items WHERE id = $1", int(menuItemID))
	if err != nil {
		log.Printf("Error deleting menu item: %v", err)
		return h.errorResponse(id, -32603, "Database error")
	}
	
	return h.successResponse(id, fmt.Sprintf("Menu item %d deleted", int(menuItemID)))
}

// Order CRUD
type Order struct {
	ID           int     `json:"id"`
	RestaurantID int     `json:"restaurant_id"`
	CustomerName string  `json:"customer_name"`
	Status       string  `json:"status"`
	TotalAmount  float64 `json:"total_amount"`
}

func (h *MCPHandler) toolListOrders(id interface{}) MCPResponse {
	rows, err := h.db.Query(`
		SELECT id, restaurant_id, customer_name, status, final_amount
		FROM orders 
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Printf("Error listing orders: %v", err)
		return h.errorResponse(id, -32603, "Database error")
	}
	defer rows.Close()
	
	orders := []Order{}
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.RestaurantID, &o.CustomerName, &o.Status, &o.TotalAmount); err != nil {
			continue
		}
		orders = append(orders, o)
	}
	
	data, _ := json.MarshalIndent(orders, "", "  ")
	return h.successResponseText(id, string(data))
}

func (h *MCPHandler) toolGetOrder(id interface{}, args map[string]interface{}) MCPResponse {
	orderID, ok := args["id"].(float64)
	if !ok {
		return h.errorResponse(id, -32602, "Missing id")
	}
	
	var order Order
	err := h.db.QueryRow(`
		SELECT id, restaurant_id, customer_name, status, final_amount
		FROM orders WHERE id = $1
	`, int(orderID)).Scan(&order.ID, &order.RestaurantID, &order.CustomerName, &order.Status, &order.TotalAmount)
	
	if err == sql.ErrNoRows {
		return h.errorResponse(id, -32602, "Order not found")
	}
	if err != nil {
		log.Printf("Error getting order: %v", err)
		return h.errorResponse(id, -32603, "Database error")
	}
	
	data, _ := json.MarshalIndent(order, "", "  ")
	return h.successResponseText(id, string(data))
}

func (h *MCPHandler) toolCreateOrder(id interface{}, args map[string]interface{}) MCPResponse {
	restaurantID, _ := args["restaurant_id"].(float64)
	customerName, _ := args["customer_name"].(string)
	items, _ := args["items"].([]interface{})
	
	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		return h.errorResponse(id, -32603, "Database error")
	}
	defer tx.Rollback()
	
	// Create order
	var orderID int
	err = tx.QueryRow(`
		INSERT INTO orders (restaurant_id, customer_name, status, total_amount, final_amount)
		VALUES ($1, $2, 'pending', 0, 0) RETURNING id
	`, int(restaurantID), customerName).Scan(&orderID)
	
	if err != nil {
		log.Printf("Error creating order: %v", err)
		return h.errorResponse(id, -32603, "Database error")
	}
	
	// Add order items
	var totalAmount float64
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		menuItemID, _ := itemMap["menu_item_id"].(float64)
		quantity, _ := itemMap["quantity"].(float64)
		
		// Get price
		var price float64
		tx.QueryRow("SELECT price FROM menu_items WHERE id = $1", int(menuItemID)).Scan(&price)
		
		// Insert order item
		tx.Exec(`
			INSERT INTO order_items (order_id, menu_item_id, quantity, price)
			VALUES ($1, $2, $3, $4)
		`, orderID, int(menuItemID), int(quantity), price)
		
		totalAmount += price * quantity
	}
	
	// Update order total
	tx.Exec("UPDATE orders SET total_amount = $1, final_amount = $1 WHERE id = $2", totalAmount, orderID)
	
	if err := tx.Commit(); err != nil {
		return h.errorResponse(id, -32603, "Database error")
	}
	
	return h.successResponse(id, fmt.Sprintf("Order created with ID %d, total: $%.2f", orderID, totalAmount))
}

func (h *MCPHandler) toolUpdateOrder(id interface{}, args map[string]interface{}) MCPResponse {
	orderID, ok := args["id"].(float64)
	if !ok {
		return h.errorResponse(id, -32602, "Missing id")
	}
	
	status, _ := args["status"].(string)
	
	_, err := h.db.Exec("UPDATE orders SET status = $1 WHERE id = $2", status, int(orderID))
	if err != nil {
		log.Printf("Error updating order: %v", err)
		return h.errorResponse(id, -32603, "Database error")
	}
	
	return h.successResponse(id, fmt.Sprintf("Order %d status updated to %s", int(orderID), status))
}

func (h *MCPHandler) toolDeleteOrder(id interface{}, args map[string]interface{}) MCPResponse {
	orderID, ok := args["id"].(float64)
	if !ok {
		return h.errorResponse(id, -32602, "Missing id")
	}
	
	_, err := h.db.Exec("DELETE FROM orders WHERE id = $1", int(orderID))
	if err != nil {
		log.Printf("Error deleting order: %v", err)
		return h.errorResponse(id, -32603, "Database error")
	}
	
	return h.successResponse(id, fmt.Sprintf("Order %d deleted", int(orderID)))
}

// Helper functions
func (h *MCPHandler) successResponse(id interface{}, message string) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": message,
				},
			},
		},
		ID: id,
	}
}

func (h *MCPHandler) successResponseText(id interface{}, text string) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": text,
				},
			},
		},
		ID: id,
	}
}
