package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/vishalk17/mcp-service-restaurant/internal/models"
	"github.com/vishalk17/mcp-service-restaurant/internal/storage"
)

// JSON-RPC 2.0 structures
type JSONRPCRequest struct {
	JsonRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP Protocol structures
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ClientCapabilities     `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

type ClientCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Sampling     map[string]interface{} `json:"sampling,omitempty"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

type ServerCapabilities struct {
	Tools        *ToolsCapability       `json:"tools,omitempty"`
	Experimental map[string]interface{} `json:"experimental,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type CallToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type MCPServer struct {
	db          *storage.DB
	initialized bool
	mu          sync.RWMutex
}

func NewMCPServer(db *storage.DB) *MCPServer {
	return &MCPServer{
		db: db,
	}
}

func (s *MCPServer) handleRequest(req JSONRPCRequest) JSONRPCResponse {
	log.Printf("Received request: method=%s id=%v", req.Method, req.ID)

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req.ID, req.Params)
	case "notifications/initialized":
		log.Println("Client initialized")
		return JSONRPCResponse{} // No response for notifications
	case "tools/list":
		s.mu.RLock()
		initialized := s.initialized
		s.mu.RUnlock()
		if !initialized {
			return s.sendError(req.ID, -32002, "Server not initialized", nil)
		}
		return s.handleToolsList(req.ID)
	case "tools/call":
		s.mu.RLock()
		initialized := s.initialized
		s.mu.RUnlock()
		if !initialized {
			return s.sendError(req.ID, -32002, "Server not initialized", nil)
		}
		return s.handleCallTool(req.ID, req.Params)
	case "ping":
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]string{},
		}
	default:
		return s.sendError(req.ID, -32601, "Method not found", req.Method)
	}
}

func (s *MCPServer) sendError(id interface{}, code int, message string, data interface{}) JSONRPCResponse {
	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

func (s *MCPServer) handleInitialize(id interface{}, params json.RawMessage) JSONRPCResponse {
	var initParams InitializeParams
	if err := json.Unmarshal(params, &initParams); err != nil {
		return s.sendError(id, -32602, "Invalid params", err.Error())
	}

	log.Printf("Initialize request from client: %s %s", initParams.ClientInfo.Name, initParams.ClientInfo.Version)

	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{},
		},
		ServerInfo: ServerInfo{
			Name:    "restaurant-mcp-server",
			Version: "1.0.0",
		},
	}

	s.mu.Lock()
	s.initialized = true
	s.mu.Unlock()

	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

func (s *MCPServer) handleToolsList(id interface{}) JSONRPCResponse {
	tools := []Tool{
		{
			Name:        "get_restaurants",
			Description: "Get a list of all Indian restaurants with their details including name, address, phone number, and cuisine type",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		{
			Name:        "get_restaurant",
			Description: "Get details of a specific restaurant by ID",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"restaurant_id": {
						Type:        "integer",
						Description: "The ID of the restaurant to retrieve",
					},
				},
				Required: []string{"restaurant_id"},
			},
		},
		{
			Name:        "get_menu",
			Description: "Get the menu items for a specific restaurant, including Indian dishes with dietary preferences and spice levels",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"restaurant_id": {
						Type:        "integer",
						Description: "The ID of the restaurant whose menu to retrieve",
					},
				},
				Required: []string{"restaurant_id"},
			},
		},
		{
			Name:        "create_restaurant",
			Description: "Create a new restaurant with details",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name": {
						Type:        "string",
						Description: "Name of the restaurant",
					},
					"address": {
						Type:        "string",
						Description: "Address of the restaurant",
					},
					"phone_number": {
						Type:        "string",
						Description: "Phone number of the restaurant",
					},
					"cuisine_type": {
						Type:        "string",
						Description: "Type of cuisine (defaults to Indian)",
					},
				},
				Required: []string{"name", "address"},
			},
		},
		{
			Name:        "update_restaurant",
			Description: "Update an existing restaurant's details",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"restaurant_id": {
						Type:        "integer",
						Description: "ID of the restaurant to update",
					},
					"name": {
						Type:        "string",
						Description: "Name of the restaurant",
					},
					"address": {
						Type:        "string",
						Description: "Address of the restaurant",
					},
					"phone_number": {
						Type:        "string",
						Description: "Phone number of the restaurant",
					},
					"cuisine_type": {
						Type:        "string",
						Description: "Type of cuisine",
					},
				},
				Required: []string{"restaurant_id", "name", "address"},
			},
		},
		{
			Name:        "delete_restaurant",
			Description: "Delete a restaurant by ID",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"restaurant_id": {
						Type:        "integer",
						Description: "ID of the restaurant to delete",
					},
				},
				Required: []string{"restaurant_id"},
			},
		},
		{
			Name:        "get_orders",
			Description: "Get a list of all orders with their details including customer info, items, billing, and payment status",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		{
			Name:        "get_order",
			Description: "Get details of a specific order by ID",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"order_id": {
						Type:        "integer",
						Description: "The ID of the order to retrieve",
					},
				},
				Required: []string{"order_id"},
			},
		},
		{
			Name:        "create_menu_item",
			Description: "Create a new menu item for a restaurant",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"restaurant_id": {
						Type:        "integer",
						Description: "ID of the restaurant",
					},
					"name": {
						Type:        "string",
						Description: "Name of the menu item",
					},
					"description": {
						Type:        "string",
						Description: "Description of the menu item",
					},
					"price": {
						Type:        "number",
						Description: "Price of the menu item",
					},
					"category": {
						Type:        "string",
						Description: "Category (appetizer, main, dessert, beverage)",
					},
					"dietary_type": {
						Type:        "string",
						Description: "Dietary type (vegetarian, non_vegetarian, vegan, jain_friendly)",
					},
					"spice_level": {
						Type:        "string",
						Description: "Spice level (mild, medium, hot, extra_hot)",
					},
					"is_available": {
						Type:        "string",
						Description: "true or false for availability",
					},
				},
				Required: []string{"restaurant_id", "name", "price"},
			},
		},
		{
			Name:        "update_menu_item",
			Description: "Update an existing menu item's details or price",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"menu_item_id": {
						Type:        "integer",
						Description: "ID of the menu item to update",
					},
					"name": {
						Type:        "string",
						Description: "Name of the menu item",
					},
					"description": {
						Type:        "string",
						Description: "Description of the menu item",
					},
					"price": {
						Type:        "number",
						Description: "Price of the menu item",
					},
					"category": {
						Type:        "string",
						Description: "Category (appetizer, main, dessert, beverage)",
					},
					"dietary_type": {
						Type:        "string",
						Description: "Dietary type (vegetarian, non_vegetarian, vegan, jain_friendly)",
					},
					"spice_level": {
						Type:        "string",
						Description: "Spice level (mild, medium, hot, very_hot)",
					},
					"is_available": {
						Type:        "string",
						Description: "true or false for availability",
					},
				},
				Required: []string{"menu_item_id"},
			},
		},
		{
			Name:        "delete_menu_item",
			Description: "Delete a menu item by ID",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"menu_item_id": {
						Type:        "integer",
						Description: "ID of the menu item to delete",
					},
				},
				Required: []string{"menu_item_id"},
			},
		},
		{
			Name:        "create_order",
			Description: "Create a new order with items, customer details, and payment information. GST tax (5%) will be automatically calculated.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"restaurant_id": {
						Type:        "integer",
						Description: "ID of the restaurant",
					},
					"customer_name": {
						Type:        "string",
						Description: "Name of the customer",
					},
					"customer_phone": {
						Type:        "string",
						Description: "Phone number of the customer",
					},
					"items": {
						Type:        "string",
						Description: "JSON string array of order items, each with menu_item_id (integer), quantity (integer), price (number), and optional notes (string)",
					},
					"discount": {
						Type:        "number",
						Description: "Discount amount (optional, defaults to 0)",
					},
					"payment_method": {
						Type:        "string",
						Description: "Payment method",
						Enum:        []string{"cash", "card", "upi", "digital_wallet"},
					},
					"billing_address": {
						Type:        "string",
						Description: "Billing address",
					},
				},
				Required: []string{"restaurant_id", "customer_name", "items"},
			},
		},
		{
			Name:        "update_order",
			Description: "Update order status or payment information",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"order_id": {
						Type:        "integer",
						Description: "ID of the order to update",
					},
					"status": {
						Type:        "string",
						Description: "Order status (pending, confirmed, preparing, ready, delivered, cancelled)",
					},
					"payment_status": {
						Type:        "string",
						Description: "Payment status (pending, completed, failed, refunded)",
					},
				},
				Required: []string{"order_id"},
			},
		},
		{
			Name:        "delete_order",
			Description: "Delete an order by ID",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"order_id": {
						Type:        "integer",
						Description: "ID of the order to delete",
					},
				},
				Required: []string{"order_id"},
			},
		},
	}

	result := ToolsListResult{Tools: tools}

	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

func (s *MCPServer) handleCallTool(id interface{}, params json.RawMessage) JSONRPCResponse {
	var callParams CallToolParams
	if err := json.Unmarshal(params, &callParams); err != nil {
		return s.sendError(id, -32602, "Invalid params", err.Error())
	}

	log.Printf("Tool call: %s with args: %v", callParams.Name, callParams.Arguments)

	switch callParams.Name {
	case "get_restaurants":
		return s.handleGetRestaurants(id)
	case "get_restaurant":
		return s.handleGetRestaurant(id, callParams.Arguments)
	case "create_restaurant":
		return s.handleCreateRestaurant(id, callParams.Arguments)
	case "update_restaurant":
		return s.handleUpdateRestaurant(id, callParams.Arguments)
	case "delete_restaurant":
		return s.handleDeleteRestaurant(id, callParams.Arguments)
	case "get_menu":
		return s.handleGetMenu(id, callParams.Arguments)
	case "create_menu_item":
		return s.handleCreateMenuItem(id, callParams.Arguments)
	case "update_menu_item":
		return s.handleUpdateMenuItem(id, callParams.Arguments)
	case "delete_menu_item":
		return s.handleDeleteMenuItem(id, callParams.Arguments)
	case "get_orders":
		return s.handleGetOrders(id)
	case "get_order":
		return s.handleGetOrder(id, callParams.Arguments)
	case "create_order":
		return s.handleCreateOrder(id, callParams.Arguments)
	case "update_order":
		return s.handleUpdateOrder(id, callParams.Arguments)
	case "delete_order":
		return s.handleDeleteOrder(id, callParams.Arguments)
	default:
		return s.sendError(id, -32601, "Unknown tool", callParams.Name)
	}
}

func (s *MCPServer) handleGetRestaurants(id interface{}) JSONRPCResponse {
	restaurants, err := s.db.GetAllRestaurants()
	if err != nil {
		log.Printf("Error getting restaurants: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	data, _ := json.MarshalIndent(restaurants, "", "  ")
	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: string(data)}},
		},
	}
}

func (s *MCPServer) handleGetRestaurant(id interface{}, args map[string]interface{}) JSONRPCResponse {
	restaurantID, ok := args["restaurant_id"].(float64)
	if !ok {
		return s.sendError(id, -32602, "Missing or invalid restaurant_id", nil)
	}

	restaurant, err := s.db.GetRestaurantByID(int(restaurantID))
	if err != nil {
		log.Printf("Error getting restaurant: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	data, _ := json.MarshalIndent(restaurant, "", "  ")
	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: string(data)}},
		},
	}
}

func (s *MCPServer) handleGetMenu(id interface{}, args map[string]interface{}) JSONRPCResponse {
	restaurantID, ok := args["restaurant_id"].(float64)
	if !ok {
		return s.sendError(id, -32602, "Missing or invalid restaurant_id", nil)
	}

	menuItems, err := s.db.GetMenuByRestaurantID(int(restaurantID))
	if err != nil {
		log.Printf("Error getting menu: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	data, _ := json.MarshalIndent(menuItems, "", "  ")
	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: string(data)}},
		},
	}
}

func (s *MCPServer) handleCreateMenuItem(id interface{}, args map[string]interface{}) JSONRPCResponse {
	restaurantID, ok := args["restaurant_id"].(float64)
	if !ok {
		return s.sendError(id, -32602, "Missing or invalid restaurant_id", nil)
	}

	name, _ := args["name"].(string)
	description, _ := args["description"].(string)
	price, ok := args["price"].(float64)
	category, _ := args["category"].(string)
	dietaryType, _ := args["dietary_type"].(string)
	spiceLevel, _ := args["spice_level"].(string)
	isAvailStr, _ := args["is_available"].(string)

	if name == "" || !ok {
		return s.sendError(id, -32602, "Missing required fields: name and price", nil)
	}

	isAvailable := true
	if isAvailStr == "false" {
		isAvailable = false
	}

	if category == "" {
		category = "Main Course"
	}

	if dietaryType == "" {
		dietaryType = "vegetarian"
	}

	if spiceLevel == "" {
		spiceLevel = "medium"
	}

	menuItem := &models.MenuItem{
		RestaurantID: int(restaurantID),
		Name:         name,
		Description:  description,
		Price:        price,
		Category:     category,
		DietaryType:  dietaryType,
		SpiceLevel:   spiceLevel,
		Available:    isAvailable,
	}

	err := s.db.CreateMenuItem(menuItem)
	if err != nil {
		log.Printf("Error creating menu item: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	data, _ := json.MarshalIndent(menuItem, "", "  ")
	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Menu item created successfully:\n%s", string(data))}},
		},
	}
}

func (s *MCPServer) handleUpdateMenuItem(id interface{}, args map[string]interface{}) JSONRPCResponse {
	menuItemID, ok := args["menu_item_id"].(float64)
	if !ok {
		return s.sendError(id, -32602, "Missing or invalid menu_item_id", nil)
	}

	// Get existing menu item first
	existingItem, err := s.db.GetMenuItemByID(int(menuItemID))
	if err != nil {
		log.Printf("Error getting menu item: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	// Update fields if provided
	if name, ok := args["name"].(string); ok && name != "" {
		existingItem.Name = name
	}
	if description, ok := args["description"].(string); ok {
		existingItem.Description = description
	}
	if price, ok := args["price"].(float64); ok {
		existingItem.Price = price
	}
	if category, ok := args["category"].(string); ok && category != "" {
		existingItem.Category = category
	}
	if dietaryType, ok := args["dietary_type"].(string); ok && dietaryType != "" {
		existingItem.DietaryType = dietaryType
	}
	if spiceLevel, ok := args["spice_level"].(string); ok && spiceLevel != "" {
		existingItem.SpiceLevel = spiceLevel
	}
	if isAvailStr, ok := args["is_available"].(string); ok {
		existingItem.Available = (isAvailStr == "true")
	}

	err = s.db.UpdateMenuItem(existingItem)
	if err != nil {
		log.Printf("Error updating menu item: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	data, _ := json.MarshalIndent(existingItem, "", "  ")
	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Menu item updated successfully:\n%s", string(data))}},
		},
	}
}

func (s *MCPServer) handleDeleteMenuItem(id interface{}, args map[string]interface{}) JSONRPCResponse {
	menuItemID, ok := args["menu_item_id"].(float64)
	if !ok {
		return s.sendError(id, -32602, "Missing or invalid menu_item_id", nil)
	}

	err := s.db.DeleteMenuItem(int(menuItemID))
	if err != nil {
		log.Printf("Error deleting menu item: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Menu item ID %d deleted successfully", int(menuItemID))}},
		},
	}
}

func (s *MCPServer) handleCreateRestaurant(id interface{}, args map[string]interface{}) JSONRPCResponse {
	name, _ := args["name"].(string)
	address, _ := args["address"].(string)
	phoneNumber, _ := args["phone_number"].(string)
	cuisineType, _ := args["cuisine_type"].(string)

	if name == "" || address == "" {
		return s.sendError(id, -32602, "Missing required fields: name and address", nil)
	}

	if cuisineType == "" {
		cuisineType = "Indian"
	}

	restaurant := &models.Restaurant{
		Name:        name,
		Address:     address,
		PhoneNumber: phoneNumber,
		CuisineType: cuisineType,
	}

	err := s.db.CreateRestaurant(restaurant)
	if err != nil {
		log.Printf("Error creating restaurant: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	data, _ := json.MarshalIndent(restaurant, "", "  ")
	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Restaurant created successfully:\n%s", string(data))}},
		},
	}
}

func (s *MCPServer) handleUpdateRestaurant(id interface{}, args map[string]interface{}) JSONRPCResponse {
	restaurantID, ok := args["restaurant_id"].(float64)
	if !ok {
		return s.sendError(id, -32602, "Missing or invalid restaurant_id", nil)
	}

	name, _ := args["name"].(string)
	address, _ := args["address"].(string)
	phoneNumber, _ := args["phone_number"].(string)
	cuisineType, _ := args["cuisine_type"].(string)

	if name == "" || address == "" {
		return s.sendError(id, -32602, "Missing required fields: name and address", nil)
	}

	restaurant := &models.Restaurant{
		ID:          int(restaurantID),
		Name:        name,
		Address:     address,
		PhoneNumber: phoneNumber,
		CuisineType: cuisineType,
	}

	err := s.db.UpdateRestaurant(restaurant)
	if err != nil {
		log.Printf("Error updating restaurant: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	data, _ := json.MarshalIndent(restaurant, "", "  ")
	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Restaurant updated successfully:\n%s", string(data))}},
		},
	}
}

func (s *MCPServer) handleDeleteRestaurant(id interface{}, args map[string]interface{}) JSONRPCResponse {
	restaurantID, ok := args["restaurant_id"].(float64)
	if !ok {
		return s.sendError(id, -32602, "Missing or invalid restaurant_id", nil)
	}

	err := s.db.DeleteRestaurant(int(restaurantID))
	if err != nil {
		log.Printf("Error deleting restaurant: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Restaurant ID %d deleted successfully", int(restaurantID))}},
		},
	}
}

func (s *MCPServer) handleGetOrders(id interface{}) JSONRPCResponse {
	orders, err := s.db.GetAllOrders()
	if err != nil {
		log.Printf("Error getting orders: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	data, _ := json.MarshalIndent(orders, "", "  ")
	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: string(data)}},
		},
	}
}

func (s *MCPServer) handleGetOrder(id interface{}, args map[string]interface{}) JSONRPCResponse {
	orderID, ok := args["order_id"].(float64)
	if !ok {
		return s.sendError(id, -32602, "Missing or invalid order_id", nil)
	}

	order, err := s.db.GetOrderByID(int(orderID))
	if err != nil {
		log.Printf("Error getting order: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	data, _ := json.MarshalIndent(order, "", "  ")
	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: string(data)}},
		},
	}
}

func (s *MCPServer) handleCreateOrder(id interface{}, args map[string]interface{}) JSONRPCResponse {
	restaurantID, ok := args["restaurant_id"].(float64)
	if !ok {
		return s.sendError(id, -32602, "Missing or invalid restaurant_id", nil)
	}

	customerName, _ := args["customer_name"].(string)
	if customerName == "" {
		return s.sendError(id, -32602, "Missing customer_name", nil)
	}

	itemsRaw, ok := args["items"].([]interface{})
	if !ok || len(itemsRaw) == 0 {
		return s.sendError(id, -32602, "Missing or invalid items array", nil)
	}

	customerPhone, _ := args["customer_phone"].(string)
	discount, _ := args["discount"].(float64)
	paymentMethod, _ := args["payment_method"].(string)
	billingAddress, _ := args["billing_address"].(string)

	if paymentMethod == "" {
		paymentMethod = "cash"
	}

	order := &models.Order{
		RestaurantID:   int(restaurantID),
		CustomerName:   customerName,
		CustomerPhone:  customerPhone,
		Status:         "pending",
		Discount:       discount,
		PaymentStatus:  "pending",
		PaymentMethod:  paymentMethod,
		BillingAddress: billingAddress,
		OrderItems:     []models.OrderItem{},
	}

	totalAmount := 0.0
	for _, itemRaw := range itemsRaw {
		itemMap, ok := itemRaw.(map[string]interface{})
		if !ok {
			continue
		}

		menuItemID, _ := itemMap["menu_item_id"].(float64)
		quantity, _ := itemMap["quantity"].(float64)
		price, _ := itemMap["price"].(float64)
		notes, _ := itemMap["notes"].(string)

		if menuItemID == 0 || quantity == 0 || price == 0 {
			continue
		}

		subtotal := float64(quantity) * price
		totalAmount += subtotal

		order.OrderItems = append(order.OrderItems, models.OrderItem{
			MenuItemID: int(menuItemID),
			Quantity:   int(quantity),
			Price:      price,
			Notes:      notes,
			Subtotal:   subtotal,
		})
	}

	order.TotalAmount = totalAmount
	order.TaxAmount = totalAmount * 0.05
	order.FinalAmount = totalAmount + order.TaxAmount - order.Discount

	err := s.db.CreateOrder(order)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	completeOrder, err := s.db.GetOrderByID(order.ID)
	if err != nil {
		log.Printf("Error fetching created order: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Order created but error fetching details: %v", err)}},
				IsError: true,
			},
		}
	}

	data, _ := json.MarshalIndent(completeOrder, "", "  ")
	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Order created successfully:\n%s", string(data))}},
		},
	}
}

func (s *MCPServer) handleUpdateOrder(id interface{}, args map[string]interface{}) JSONRPCResponse {
	orderID, ok := args["order_id"].(float64)
	if !ok {
		return s.sendError(id, -32602, "Missing or invalid order_id", nil)
	}

	// Get existing order first
	existingOrder, err := s.db.GetOrderByID(int(orderID))
	if err != nil {
		log.Printf("Error getting order: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	// Update fields if provided
	if status, ok := args["status"].(string); ok && status != "" {
		existingOrder.Status = status
	}
	if paymentStatus, ok := args["payment_status"].(string); ok && paymentStatus != "" {
		existingOrder.PaymentStatus = paymentStatus
	}

	err = s.db.UpdateOrder(existingOrder)
	if err != nil {
		log.Printf("Error updating order: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	data, _ := json.MarshalIndent(existingOrder, "", "  ")
	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Order updated successfully:\n%s", string(data))}},
		},
	}
}

func (s *MCPServer) handleDeleteOrder(id interface{}, args map[string]interface{}) JSONRPCResponse {
	orderID, ok := args["order_id"].(float64)
	if !ok {
		return s.sendError(id, -32602, "Missing or invalid order_id", nil)
	}

	err := s.db.DeleteOrder(int(orderID))
	if err != nil {
		log.Printf("Error deleting order: %v", err)
		return JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	return JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Order ID %d deleted successfully", int(orderID))}},
		},
	}
}

// SSE Handler for remote MCP
func (s *MCPServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Printf("SSE connection from %s", r.RemoteAddr)

	// Create a channel for this client
	messageChan := make(chan string, 10)
	defer close(messageChan)

	// Handle POST body as JSON-RPC request
	if r.Method == "POST" {
		var req JSONRPCRequest
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&req); err != nil {
			log.Printf("Error decoding request: %v", err)
			http.Error(w, "Invalid JSON-RPC request", http.StatusBadRequest)
			return
		}

		// Process the request
		response := s.handleRequest(req)

		// Send response as SSE
		if response.JsonRPC != "" { // Don't send empty responses for notifications
			data, _ := json.Marshal(response)
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()
		}
		return
	}

	// For GET requests, handle as streaming connection
	scanner := bufio.NewScanner(r.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			log.Printf("Error parsing request: %v", err)
			continue
		}

		response := s.handleRequest(req)
		if response.JsonRPC != "" {
			data, _ := json.Marshal(response)
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()
		}
	}
}

// Health check
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"server": "remote-mcp-server",
	})
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Get database connection string
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "host=localhost port=5432 user=postgres password=postgres dbname=mcp_restaurant sslmode=disable"
	}

	// Initialize database
	db, err := storage.NewDB(dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	// Create MCP server
	server := NewMCPServer(db)

	// Setup HTTP handlers
	http.HandleFunc("/mcp", server.handleSSE)
	http.HandleFunc("/health", healthCheck)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Remote MCP Server starting on port %s", port)
	log.Printf("Endpoint: /mcp (SSE transport)")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
