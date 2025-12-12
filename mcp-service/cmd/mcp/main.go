package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

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

type JSONRPCNotification struct {
	JsonRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
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
	Tools       *ToolsCapability       `json:"tools,omitempty"`
	Resources   *ResourcesCapability   `json:"resources,omitempty"`
	Prompts     *PromptsCapability     `json:"prompts,omitempty"`
	Logging     map[string]interface{} `json:"logging,omitempty"`
	Experimental map[string]interface{} `json:"experimental,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

type PromptsCapability struct {
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
	Type       string                 `json:"type"`
	Properties map[string]Property    `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
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
	scanner     *bufio.Scanner
	initialized bool
}

func NewMCPServer(db *storage.DB) *MCPServer {
	return &MCPServer{
		db:      db,
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (s *MCPServer) sendResponse(resp interface{}) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func (s *MCPServer) sendError(id interface{}, code int, message string, data interface{}) error {
	return s.sendResponse(JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	})
}

func (s *MCPServer) handleInitialize(id interface{}, params json.RawMessage) error {
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

	s.initialized = true

	return s.sendResponse(JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func (s *MCPServer) handleToolsList(id interface{}) error {
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
						Description: "Type of cuisine (e.g., Indian, North Indian, South Indian)",
					},
				},
				Required: []string{"name", "address"},
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
						Type:        "array",
						Description: "Array of order items with menu_item_id, quantity, price, and optional notes",
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
	}

	result := ToolsListResult{Tools: tools}

	return s.sendResponse(JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func (s *MCPServer) handleCallTool(id interface{}, params json.RawMessage) error {
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
	case "get_menu":
		return s.handleGetMenu(id, callParams.Arguments)
	case "create_restaurant":
		return s.handleCreateRestaurant(id, callParams.Arguments)
	case "get_orders":
		return s.handleGetOrders(id)
	case "get_order":
		return s.handleGetOrder(id, callParams.Arguments)
	case "create_order":
		return s.handleCreateOrder(id, callParams.Arguments)
	default:
		return s.sendError(id, -32601, "Unknown tool", callParams.Name)
	}
}

func (s *MCPServer) handleGetRestaurants(id interface{}) error {
	restaurants, err := s.db.GetAllRestaurants()
	if err != nil {
		log.Printf("Error getting restaurants: %v", err)
		return s.sendResponse(JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		})
	}

	data, _ := json.MarshalIndent(restaurants, "", "  ")
	return s.sendResponse(JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: string(data)}},
		},
	})
}

func (s *MCPServer) handleGetRestaurant(id interface{}, args map[string]interface{}) error {
	restaurantID, ok := args["restaurant_id"].(float64)
	if !ok {
		return s.sendError(id, -32602, "Missing or invalid restaurant_id", nil)
	}

	restaurant, err := s.db.GetRestaurantByID(int(restaurantID))
	if err != nil {
		log.Printf("Error getting restaurant: %v", err)
		return s.sendResponse(JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		})
	}

	data, _ := json.MarshalIndent(restaurant, "", "  ")
	return s.sendResponse(JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: string(data)}},
		},
	})
}

func (s *MCPServer) handleGetMenu(id interface{}, args map[string]interface{}) error {
	restaurantID, ok := args["restaurant_id"].(float64)
	if !ok {
		return s.sendError(id, -32602, "Missing or invalid restaurant_id", nil)
	}

	menuItems, err := s.db.GetMenuByRestaurantID(int(restaurantID))
	if err != nil {
		log.Printf("Error getting menu: %v", err)
		return s.sendResponse(JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		})
	}

	data, _ := json.MarshalIndent(menuItems, "", "  ")
	return s.sendResponse(JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: string(data)}},
		},
	})
}

func (s *MCPServer) handleCreateRestaurant(id interface{}, args map[string]interface{}) error {
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
		return s.sendResponse(JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		})
	}

	data, _ := json.MarshalIndent(restaurant, "", "  ")
	return s.sendResponse(JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Restaurant created successfully:\n%s", string(data))}},
		},
	})
}

func (s *MCPServer) handleGetOrders(id interface{}) error {
	orders, err := s.db.GetAllOrders()
	if err != nil {
		log.Printf("Error getting orders: %v", err)
		return s.sendResponse(JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		})
	}

	data, _ := json.MarshalIndent(orders, "", "  ")
	return s.sendResponse(JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: string(data)}},
		},
	})
}

func (s *MCPServer) handleGetOrder(id interface{}, args map[string]interface{}) error {
	orderID, ok := args["order_id"].(float64)
	if !ok {
		return s.sendError(id, -32602, "Missing or invalid order_id", nil)
	}

	order, err := s.db.GetOrderByID(int(orderID))
	if err != nil {
		log.Printf("Error getting order: %v", err)
		return s.sendResponse(JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		})
	}

	data, _ := json.MarshalIndent(order, "", "  ")
	return s.sendResponse(JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: string(data)}},
		},
	})
}

func (s *MCPServer) handleCreateOrder(id interface{}, args map[string]interface{}) error {
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

	// Parse order items
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

	// Calculate GST and final amount
	order.TotalAmount = totalAmount
	order.TaxAmount = totalAmount * 0.05 // 5% GST
	order.FinalAmount = totalAmount + order.TaxAmount - order.Discount

	err := s.db.CreateOrder(order)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		return s.sendResponse(JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		})
	}

	// Fetch complete order with all details
	completeOrder, err := s.db.GetOrderByID(order.ID)
	if err != nil {
		log.Printf("Error fetching created order: %v", err)
		return s.sendResponse(JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []Content{{Type: "text", Text: fmt.Sprintf("Order created but error fetching details: %v", err)}},
				IsError: true,
			},
		})
	}

	data, _ := json.MarshalIndent(completeOrder, "", "  ")
	return s.sendResponse(JSONRPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Order created successfully:\n%s", string(data))}},
		},
	})
}

func (s *MCPServer) handleRequest(line string) error {
	var req JSONRPCRequest
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		log.Printf("Invalid JSON-RPC request: %v", err)
		return s.sendError(nil, -32700, "Parse error", err.Error())
	}

	log.Printf("Received request: method=%s id=%v", req.Method, req.ID)

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req.ID, req.Params)
	case "notifications/initialized":
		// Client is telling us initialization is complete
		log.Println("Client initialized")
		return nil
	case "tools/list":
		if !s.initialized {
			return s.sendError(req.ID, -32002, "Server not initialized", nil)
		}
		return s.handleToolsList(req.ID)
	case "tools/call":
		if !s.initialized {
			return s.sendError(req.ID, -32002, "Server not initialized", nil)
		}
		return s.handleCallTool(req.ID, req.Params)
	case "ping":
		// Handle ping for testing
		return s.sendResponse(JSONRPCResponse{
			JsonRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]string{},
		})
	default:
		return s.sendError(req.ID, -32601, "Method not found", req.Method)
	}
}

func (s *MCPServer) Run() {
	log.Println("MCP Server started, listening on stdin...")

	for s.scanner.Scan() {
		line := s.scanner.Text()
		if line == "" {
			continue
		}

		log.Printf("Received: %s", line)

		if err := s.handleRequest(line); err != nil {
			log.Printf("Error handling request: %v", err)
		}
	}

	if err := s.scanner.Err(); err != nil {
		log.Fatalf("Scanner error: %v", err)
	}
}

func main() {
	// Set up logging to stderr (stdout is reserved for JSON-RPC communication)
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Get database connection string from environment variable
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

	// Create and run MCP server
	server := NewMCPServer(db)
	server.Run()
}
