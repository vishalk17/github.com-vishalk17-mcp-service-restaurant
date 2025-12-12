package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/vishalk17/mcp-service-restaurant/internal/models"
	"github.com/vishalk17/mcp-service-restaurant/internal/storage"
)

// JSON-RPC 2.0 structs
type JSONRPCRequest struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin - adjust as needed for security
		return true
	},
}

type Handlers struct {
	DB *storage.DB
}

func NewHandlers(db *storage.DB) *Handlers {
	return &Handlers{DB: db}
}

// MCP WebSocket handler for the root endpoint
func (h *Handlers) MCPWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("WebSocket connection request to /mcp")

	// Perform the WebSocket upgrade
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		// Don't send any response here as upgrader.Upgrade already handles HTTP response
		return
	}
	defer conn.Close()

	log.Println("WebSocket connection established successfully")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			} else {
				log.Println("WebSocket connection closed normally")
			}
			break
		}

		log.Printf("Received JSON-RPC message: %s", message)

		var req JSONRPCRequest
		if err := json.Unmarshal(message, &req); err != nil {
			log.Printf("Invalid JSON-RPC message: %v", err)

			// Send back a parse error response
			errorResponse := JSONRPCResponse{
				JsonRPC: "2.0",
				ID:      nil, // Parse errors typically have ID as null
				Error: &RPCError{
					Code:    -32700, // Parse error
					Message: "Parse error: Invalid JSON",
				},
			}

			responseBytes, marshalErr := json.Marshal(errorResponse)
			if marshalErr != nil {
				log.Printf("Failed to marshal error response: %v", marshalErr)
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
				log.Printf("WebSocket write error: %v", err)
				break
			}
			continue
		}

		var response JSONRPCResponse

		// Handle initialization request
		if req.Method == "initialize" {
			response = JSONRPCResponse{
				JsonRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"protocolVersion": "2024-06-01",
					"capabilities": map[string]interface{}{
						"tools":     map[string]interface{}{},
						"resources": map[string]interface{}{},
						"prompts":   map[string]interface{}{},
					},
				},
			}
		} else if req.Method == "tools/list" {
			// Return list of available tools
			response = JSONRPCResponse{
				JsonRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"tools": []map[string]interface{}{
						{
							"name":        "get_restaurants",
							"description": "Get a list of all restaurants",
						},
						{
							"name":        "get_restaurant_menu",
							"description": "Get menu for a specific restaurant",
						},
						{
							"name":        "get_orders",
							"description": "Get a list of all orders",
						},
					},
				},
			}
		} else if req.Method == "tools/call" {
			// Handle tool call
			params, ok := req.Params.(map[string]interface{})
			if !ok {
				response = JSONRPCResponse{
					JsonRPC: "2.0",
					ID:      req.ID,
					Error: &RPCError{
						Code:    -32602,
						Message: "Invalid params",
					},
				}
			} else {
				toolName, ok := params["name"].(string)
				if !ok {
					response = JSONRPCResponse{
						JsonRPC: "2.0",
						ID:      req.ID,
						Error: &RPCError{
							Code:    -32602,
							Message: "Missing tool name",
						},
					}
				} else {
					toolArgs, _ := params["arguments"].(map[string]interface{})
					response = h.handleToolCall(toolName, toolArgs, req.ID)
				}
			}
		} else {
			// For other methods, return an error
			response = JSONRPCResponse{
				JsonRPC: "2.0",
				ID:      req.ID,
				Error: &RPCError{
					Code:    -32601,
					Message: "Method not found: " + req.Method,
				},
			}
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			log.Printf("JSON marshal error: %v", err)
			continue
		}

		log.Printf("Sending JSON-RPC response: %s", responseBytes)

		if err := conn.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}

	log.Println("WebSocket connection closed")
}

// Handle tool calls for restaurant management
func (h *Handlers) handleToolCall(toolName string, args map[string]interface{}, id interface{}) JSONRPCResponse {
	switch toolName {
	case "get_restaurants":
		restaurants, err := h.DB.GetAllRestaurants()
		if err != nil {
			return JSONRPCResponse{
				JsonRPC: "2.0",
				ID:      id,
				Error: &RPCError{
					Code:    -32603,
					Message: "Internal error: " + err.Error(),
				},
			}
		}
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result:  restaurants,
		}
	case "get_restaurant_menu":
		id, exists := args["restaurant_id"]
		if !exists {
			return JSONRPCResponse{
				JsonRPC: "2.0",
				ID:      id,
				Error: &RPCError{
					Code:    -32602,
					Message: "Missing restaurant_id parameter",
				},
			}
		}
		restaurantID, ok := id.(float64) // JSON numbers are float64
		if !ok {
			return JSONRPCResponse{
				JsonRPC: "2.0",
				ID:      id,
				Error: &RPCError{
					Code:    -32602,
					Message: "Invalid restaurant_id parameter",
				},
			}
		}
		menuItems, err := h.DB.GetMenuByRestaurantID(int(restaurantID))
		if err != nil {
			return JSONRPCResponse{
				JsonRPC: "2.0",
				ID:      id,
				Error: &RPCError{
					Code:    -32603,
					Message: "Internal error: " + err.Error(),
				},
			}
		}
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result:  menuItems,
		}
	case "get_orders":
		orders, err := h.DB.GetAllOrders()
		if err != nil {
			return JSONRPCResponse{
				JsonRPC: "2.0",
				ID:      id,
				Error: &RPCError{
					Code:    -32603,
					Message: "Internal error: " + err.Error(),
				},
			}
		}
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result:  orders,
		}
	default:
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Error: &RPCError{
				Code:    -32601,
				Message: "Unknown tool: " + toolName,
			},
		}
	}
}

// Restaurant Handlers (to keep existing functionality accessible for MCP tools)
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