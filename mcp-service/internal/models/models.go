package models

import "time"

// Restaurant represents a restaurant entity
type Restaurant struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	PhoneNumber string    `json:"phone_number"`
	CuisineType string    `json:"cuisine_type"`  // Should be "Indian" for our use case
	CreatedAt   time.Time `json:"created_at"`
}

// MenuItem represents an Indian cuisine menu item
type MenuItem struct {
	ID          int     `json:"id"`
	RestaurantID int    `json:"restaurant_id"`
	Name        string  `json:"name"`          // Name of the Indian dish
	Description string  `json:"description"`   // Description of the dish
	Price       float64 `json:"price"`         // Price in INR
	Category    string  `json:"category"`      // North Indian, South Indian, Street Food, Sweets, Beverages
	DietaryType string  `json:"dietary_type"`  // vegetarian, non_vegetarian, vegan, jain_friendly
	SpiceLevel  string  `json:"spice_level"`   // mild, medium, hot, very_hot
	Available   bool    `json:"available"`
	CreatedAt   time.Time `json:"created_at"`
}

// Order represents a customer order
type Order struct {
	ID           int       `json:"id"`
	RestaurantID int       `json:"restaurant_id"`
	CustomerName string    `json:"customer_name"`
	CustomerPhone string   `json:"customer_phone"`
	OrderItems   []OrderItem `json:"order_items"`
	Status       string    `json:"status"`       // pending, cooking, ready, delivered, cancelled
	TotalAmount  float64   `json:"total_amount"` // Subtotal before tax/discount
	TaxAmount    float64   `json:"tax_amount"`   // GST amount
	Discount     float64   `json:"discount"`     // Discount amount
	FinalAmount  float64   `json:"final_amount"` // Total after tax and discount
	PaymentStatus string  `json:"payment_status"` // pending, paid, refunded
	PaymentMethod string  `json:"payment_method"` // cash, card, upi, digital_wallet
	BillingAddress string `json:"billing_address"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID         int     `json:"id"`
	OrderID    int     `json:"order_id"`
	MenuItemID int     `json:"menu_item_id"`
	MenuItem   MenuItem `json:"menu_item"`      // Embedded menu item details
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`            // Price at time of order
	Notes      string  `json:"notes"`            // Special instructions
	Subtotal   float64 `json:"subtotal"`         // Quantity * Price
}