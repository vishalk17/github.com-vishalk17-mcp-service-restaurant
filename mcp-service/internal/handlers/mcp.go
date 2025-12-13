package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	mw "github.com/vishalk17/mcp-service-restaurant/internal/middleware"
)

type MCPHandler struct {
	db *sql.DB
}

func NewMCPHandler(db *sql.DB) *MCPHandler {
	return &MCPHandler{db: db}
}

// MCP JSON-RPC types
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// HandleMCP handles MCP JSON-RPC requests
func (h *MCPHandler) HandleMCP(w http.ResponseWriter, r *http.Request) {
	if mw.IsDebug() {
		log.Printf("MCP request from %s: %s", r.RemoteAddr, r.URL.Path)
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, req.ID, -32700, "Parse error")
		return
	}

	if mw.IsDebug() {
		log.Printf("MCP method: %s", req.Method)
	}

	var response MCPResponse

	switch req.Method {
	case "initialize":
		response = h.handleInitialize(req.ID)
	case "notifications/initialized":
		if mw.IsDebug() { log.Println("Client initialized notification") }
		w.WriteHeader(http.StatusOK)
		return // No response for notifications
	case "tools/list":
		response = h.handleToolsList(req.ID)
	case "tools/call":
		response = h.handleToolsCall(req)
	default:
		response = MCPResponse{
			JSONRPC: "2.0",
			Error: &MCPError{
				Code:    -32601,
				Message: "Method not found: " + req.Method,
			},
			ID: req.ID,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *MCPHandler) handleInitialize(id interface{}) MCPResponse {
	if mw.IsDebug() {
		log.Printf("Initialize request")
	}

	return MCPResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "restaurant-mcp-server",
				"version": "1.0.0",
			},
		},
		ID: id,
	}
}

func (h *MCPHandler) handleToolsList(id interface{}) MCPResponse {
	tools := []map[string]interface{}{
		{"name": "list_restaurants", "description": "List all restaurants", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}},
		{"name": "get_restaurant", "description": "Get restaurant by ID", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"id": map[string]interface{}{"type": "number"}}, "required": []string{"id"}}},
		{"name": "create_restaurant", "description": "Create a new restaurant", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"name": map[string]interface{}{"type": "string"}, "address": map[string]interface{}{"type": "string"}, "phone_number": map[string]interface{}{"type": "string"}, "cuisine_type": map[string]interface{}{"type": "string"}}, "required": []string{"name", "address"}}},
		{"name": "update_restaurant", "description": "Update restaurant", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"id": map[string]interface{}{"type": "number"}, "name": map[string]interface{}{"type": "string"}, "address": map[string]interface{}{"type": "string"}, "phone_number": map[string]interface{}{"type": "string"}, "cuisine_type": map[string]interface{}{"type": "string"}}, "required": []string{"id"}}},
		{"name": "delete_restaurant", "description": "Delete restaurant", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"id": map[string]interface{}{"type": "number"}}, "required": []string{"id"}}},
		{"name": "get_menu", "description": "Get menu for restaurant", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"restaurant_id": map[string]interface{}{"type": "number"}}, "required": []string{"restaurant_id"}}},
		{"name": "create_menu_item", "description": "Add menu item", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"restaurant_id": map[string]interface{}{"type": "number"}, "name": map[string]interface{}{"type": "string"}, "description": map[string]interface{}{"type": "string"}, "price": map[string]interface{}{"type": "number"}, "category": map[string]interface{}{"type": "string"}, "dietary_type": map[string]interface{}{"type": "string"}, "spice_level": map[string]interface{}{"type": "string"}}, "required": []string{"restaurant_id", "name", "price"}}},
		{"name": "update_menu_item", "description": "Update menu item", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"id": map[string]interface{}{"type": "number"}, "name": map[string]interface{}{"type": "string"}, "description": map[string]interface{}{"type": "string"}, "price": map[string]interface{}{"type": "number"}, "category": map[string]interface{}{"type": "string"}}, "required": []string{"id"}}},
		{"name": "delete_menu_item", "description": "Delete menu item", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"id": map[string]interface{}{"type": "number"}}, "required": []string{"id"}}},
		{"name": "list_orders", "description": "List all orders", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}},
		{"name": "get_order", "description": "Get order by ID", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"id": map[string]interface{}{"type": "number"}}, "required": []string{"id"}}},
		{"name": "create_order", "description": "Create new order", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"restaurant_id": map[string]interface{}{"type": "number"}, "customer_name": map[string]interface{}{"type": "string"}, "items": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"menu_item_id": map[string]interface{}{"type": "number"}, "quantity": map[string]interface{}{"type": "number"}}}}}, "required": []string{"restaurant_id", "customer_name", "items"}}},
		{"name": "update_order", "description": "Update order status", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"id": map[string]interface{}{"type": "number"}, "status": map[string]interface{}{"type": "string"}}, "required": []string{"id", "status"}}},
		{"name": "delete_order", "description": "Delete order", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"id": map[string]interface{}{"type": "number"}}, "required": []string{"id"}}},
	}

	return MCPResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"tools": tools,
		},
		ID: id,
	}
}

func (h *MCPHandler) handleToolsCall(req MCPRequest) MCPResponse {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return h.errorResponse(req.ID, -32602, "Invalid params")
	}

	if mw.IsDebug() {
		log.Printf("Tool call: %s with args: %v", params.Name, params.Arguments)
	}

	switch params.Name {
	case "list_restaurants":
		return h.toolListRestaurants(req.ID)
	case "get_restaurant":
		return h.toolGetRestaurant(req.ID, params.Arguments)
	case "create_restaurant":
		return h.toolCreateRestaurant(req.ID, params.Arguments)
	case "update_restaurant":
		return h.toolUpdateRestaurant(req.ID, params.Arguments)
	case "delete_restaurant":
		return h.toolDeleteRestaurant(req.ID, params.Arguments)
	case "get_menu":
		return h.toolGetMenu(req.ID, params.Arguments)
	case "create_menu_item":
		return h.toolCreateMenuItem(req.ID, params.Arguments)
	case "update_menu_item":
		return h.toolUpdateMenuItem(req.ID, params.Arguments)
	case "delete_menu_item":
		return h.toolDeleteMenuItem(req.ID, params.Arguments)
	case "list_orders":
		return h.toolListOrders(req.ID)
	case "get_order":
		return h.toolGetOrder(req.ID, params.Arguments)
	case "create_order":
		return h.toolCreateOrder(req.ID, params.Arguments)
	case "update_order":
		return h.toolUpdateOrder(req.ID, params.Arguments)
	case "delete_order":
		return h.toolDeleteOrder(req.ID, params.Arguments)
	default:
		return h.errorResponse(req.ID, -32601, "Unknown tool: "+params.Name)
	}
}

func (h *MCPHandler) toolListRestaurants(id interface{}) MCPResponse {
	rows, err := h.db.Query(`
		SELECT id, name, address, phone_number, cuisine_type 
		FROM restaurants 
		ORDER BY name
	`)
	if err != nil {
		log.Printf("Error listing restaurants: %v", err)
		return h.errorResponse(id, -32603, "Database error")
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

	data, _ := json.MarshalIndent(restaurants, "", "  ")
	
	return MCPResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": string(data),
				},
			},
		},
		ID: id,
	}
}

func (h *MCPHandler) toolGetRestaurant(id interface{}, args map[string]interface{}) MCPResponse {
	restaurantID, ok := args["id"].(float64)
	if !ok {
		return h.errorResponse(id, -32602, "Missing or invalid id")
	}

	var restaurant Restaurant
	err := h.db.QueryRow(`
		SELECT id, name, address, phone_number, cuisine_type 
		FROM restaurants 
		WHERE id = $1
	`, int(restaurantID)).Scan(&restaurant.ID, &restaurant.Name, &restaurant.Address, &restaurant.PhoneNumber, &restaurant.CuisineType)

	if err == sql.ErrNoRows {
		return h.errorResponse(id, -32602, "Restaurant not found")
	}
	if err != nil {
		log.Printf("Error getting restaurant: %v", err)
		return h.errorResponse(id, -32603, "Database error")
	}

	data, _ := json.MarshalIndent(restaurant, "", "  ")
	
	return MCPResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": string(data),
				},
			},
		},
		ID: id,
	}
}

func (h *MCPHandler) toolGetMenu(id interface{}, args map[string]interface{}) MCPResponse {
	restaurantID, ok := args["restaurant_id"].(float64)
	if !ok {
		return h.errorResponse(id, -32602, "Missing or invalid restaurant_id")
	}

	rows, err := h.db.Query(`
		SELECT id, restaurant_id, name, description, price, category, dietary_type, spice_level, available
		FROM menu_items 
		WHERE restaurant_id = $1 AND available = true
		ORDER BY category, name
	`, int(restaurantID))
	if err != nil {
		log.Printf("Error getting menu: %v", err)
		return h.errorResponse(id, -32603, "Database error")
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

	data, _ := json.MarshalIndent(menuItems, "", "  ")
	
	return MCPResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": string(data),
				},
			},
		},
		ID: id,
	}
}

func (h *MCPHandler) sendError(w http.ResponseWriter, id interface{}, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPResponse{
		JSONRPC: "2.0",
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
		ID: id,
	})
}

func (h *MCPHandler) errorResponse(id interface{}, code int, message string) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
}
