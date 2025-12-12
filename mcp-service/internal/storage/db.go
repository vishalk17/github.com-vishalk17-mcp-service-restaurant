package storage

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/vishalk17/mcp-service-restaurant/internal/models"
	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

// NewDB initializes a new database connection
func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	database := &DB{db}

	// Initialize the schema
	if err = database.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %v", err)
	}

	// Insert sample data for Indian cuisine
	if err = database.seedSampleData(); err != nil {
		log.Printf("Failed to seed sample data: %v", err)
	}

	return database, nil
}

func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS restaurants (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		address TEXT NOT NULL,
		phone_number TEXT,
		cuisine_type TEXT DEFAULT 'Indian',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS menu_items (
		id SERIAL PRIMARY KEY,
		restaurant_id INTEGER REFERENCES restaurants(id),
		name TEXT NOT NULL,
		description TEXT,
		price DECIMAL(10, 2) NOT NULL,
		category TEXT,
		dietary_type TEXT,
		spice_level TEXT,
		available BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		restaurant_id INTEGER REFERENCES restaurants(id),
		customer_name TEXT NOT NULL,
		customer_phone TEXT,
		status TEXT DEFAULT 'pending',
		total_amount DECIMAL(10, 2) DEFAULT 0.00,
		tax_amount DECIMAL(10, 2) DEFAULT 0.00,
		discount DECIMAL(10, 2) DEFAULT 0.00,
		final_amount DECIMAL(10, 2) DEFAULT 0.00,
		payment_status TEXT DEFAULT 'pending',
		payment_method TEXT,
		billing_address TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS order_items (
		id SERIAL PRIMARY KEY,
		order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
		menu_item_id INTEGER REFERENCES menu_items(id),
		quantity INTEGER NOT NULL DEFAULT 1,
		price DECIMAL(10, 2) NOT NULL,
		notes TEXT,
		subtotal DECIMAL(10, 2) GENERATED ALWAYS AS (quantity * price) STORED
	);
	`

	_, err := db.Exec(schema)
	return err
}

func (db *DB) seedSampleData() error {
	// Check if restaurants already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM restaurants").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		// Already seeded
		return nil
	}

	// Insert sample restaurants and get their IDs
	restaurantData := []struct {
		name        string
		address     string
		phoneNumber string
	}{
		{
			name:        "Taj Mahal Restaurant",
			address:     "Connaught Place, New Delhi",
			phoneNumber: "+91-11-12345678",
		},
		{
			name:        "Surya Mahal",
			address:     "Linking Road, Mumbai",
			phoneNumber: "+91-22-87654321",
		},
		{
			name:        "Hyderabad House",
			address:     "Banjara Hills, Hyderabad",
			phoneNumber: "+91-40-23456789",
		},
	}

	restaurantIDs := make([]int, len(restaurantData))
	for i, r := range restaurantData {
		err := db.QueryRow(
			"INSERT INTO restaurants (name, address, phone_number, cuisine_type) VALUES ($1, $2, $3, $4) RETURNING id",
			r.name, r.address, r.phoneNumber, "Indian",
		).Scan(&restaurantIDs[i])
		if err != nil {
			return err
		}
	}

	// Insert sample menu items for Indian cuisine
	menuItems := []struct {
		restaurantID  int
		name          string
		description   string
		price         float64
		category      string
		dietaryType   string
		spiceLevel    string
	}{
		// Taj Mahal Restaurant menu items (using actual restaurant IDs)
		{restaurantIDs[0], "Butter Chicken", "Creamy tomato-based curry with tender chicken pieces", 350.00, "Main Course", "non_vegetarian", "medium"},
		{restaurantIDs[0], "Dal Makhani", "Slow-cooked black lentils with butter and cream", 280.00, "Main Course", "vegetarian", "mild"},
		{restaurantIDs[0], "Paneer Tikka Masala", "Grilled paneer in rich, spiced tomato gravy", 320.00, "Main Course", "vegetarian", "medium"},
		{restaurantIDs[0], "Chicken Biryani", "Fragrant basmati rice with marinated chicken", 380.00, "Main Course", "non_vegetarian", "medium"},
		{restaurantIDs[0], "Naan", "Traditional leavened flatbread cooked in tandoor", 60.00, "Bread", "vegetarian", "mild"},
		{restaurantIDs[0], "Gulab Jamun", "Deep-fried milk-solid balls soaked in sugar syrup", 120.00, "Dessert", "vegetarian", "mild"},

		// Surya Mahal menu items (using actual restaurant IDs)
		{restaurantIDs[1], "Masala Dosa", "Crispy fermented crepe filled with spiced potatoes", 180.00, "South Indian", "vegetarian", "medium"},
		{restaurantIDs[1], "Idli Sambar", "Steamed rice cakes served with lentil curry", 150.00, "South Indian", "vegetarian", "medium"},
		{restaurantIDs[1], "Vada Pav", "Deep-fried potato dumpling in soft bread rolls", 80.00, "Street Food", "vegetarian", "hot"},
		{restaurantIDs[1], "Pav Bhaji", "Spiced vegetable curry served with buttered bread", 220.00, "Street Food", "vegetarian", "hot"},
		{restaurantIDs[1], "Filter Coffee", "Traditional South Indian coffee with chicory", 60.00, "Beverages", "vegetarian", "mild"},
		{restaurantIDs[1], "Mysore Pak", "Rich sweet made from gram flour, ghee, and sugar", 150.00, "Sweets", "vegetarian", "mild"},

		// Hyderabad House menu items (using actual restaurant IDs)
		{restaurantIDs[2], "Hyderabadi Biryani", "Famous aromatic rice dish with meat and spices", 360.00, "Main Course", "non_vegetarian", "hot"},
		{restaurantIDs[2], "Haleem", "Slow-cooked stew of wheat, barley, meat and lentils", 450.00, "Main Course", "non_vegetarian", "medium"},
		{restaurantIDs[2], "Chicken 65", "Spicy deep-fried chicken curry", 340.00, "Starters", "non_vegetarian", "hot"},
		{restaurantIDs[2], "Mirchi Ka Salan", "Cooked green chilies in peanut and sesame gravy", 280.00, "Side Dish", "vegetarian", "hot"},
		{restaurantIDs[2], "Double Ka Meetha", "Hyderabadi bread pudding", 180.00, "Dessert", "vegetarian", "mild"},
		{restaurantIDs[2], "Lassi", "Traditional yogurt-based drink", 100.00, "Beverages", "vegetarian", "mild"},
	}

	for _, m := range menuItems {
		_, err := db.Exec(
			"INSERT INTO menu_items (restaurant_id, name, description, price, category, dietary_type, spice_level) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			m.restaurantID, m.name, m.description, m.price, m.category, m.dietaryType, m.spiceLevel,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// Restaurant methods
func (db *DB) GetAllRestaurants() ([]models.Restaurant, error) {
	rows, err := db.Query("SELECT id, name, address, phone_number, cuisine_type, created_at FROM restaurants ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var restaurants []models.Restaurant
	for rows.Next() {
		var restaurant models.Restaurant
		err := rows.Scan(&restaurant.ID, &restaurant.Name, &restaurant.Address, &restaurant.PhoneNumber, &restaurant.CuisineType, &restaurant.CreatedAt)
		if err != nil {
			return nil, err
		}
		restaurants = append(restaurants, restaurant)
	}

	return restaurants, nil
}

func (db *DB) GetRestaurantByID(id int) (*models.Restaurant, error) {
	var restaurant models.Restaurant
	err := db.QueryRow(
		"SELECT id, name, address, phone_number, cuisine_type, created_at FROM restaurants WHERE id = $1",
		id,
	).Scan(&restaurant.ID, &restaurant.Name, &restaurant.Address, &restaurant.PhoneNumber, &restaurant.CuisineType, &restaurant.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("restaurant with ID %d not found", id)
		}
		return nil, err
	}

	return &restaurant, nil
}

func (db *DB) CreateRestaurant(restaurant *models.Restaurant) error {
	query := `INSERT INTO restaurants (name, address, phone_number, cuisine_type) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	err := db.QueryRow(query, restaurant.Name, restaurant.Address, restaurant.PhoneNumber, restaurant.CuisineType).Scan(&restaurant.ID, &restaurant.CreatedAt)
	return err
}

func (db *DB) UpdateRestaurant(id int, restaurant *models.Restaurant) error {
	query := `UPDATE restaurants SET name = $1, address = $2, phone_number = $3, cuisine_type = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $5 RETURNING id, created_at`
	err := db.QueryRow(query, restaurant.Name, restaurant.Address, restaurant.PhoneNumber, restaurant.CuisineType, id).Scan(&restaurant.ID, &restaurant.CreatedAt)
	return err
}

func (db *DB) DeleteRestaurant(id int) error {
	_, err := db.Exec("DELETE FROM restaurants WHERE id = $1", id)
	return err
}

// Menu Item methods
func (db *DB) GetMenuByRestaurantID(restaurantID int) ([]models.MenuItem, error) {
	rows, err := db.Query(
		"SELECT id, restaurant_id, name, description, price, category, dietary_type, spice_level, available, created_at FROM menu_items WHERE restaurant_id = $1 AND available = true ORDER BY category, name",
		restaurantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menuItems []models.MenuItem
	for rows.Next() {
		var menuItem models.MenuItem
		err := rows.Scan(
			&menuItem.ID,
			&menuItem.RestaurantID,
			&menuItem.Name,
			&menuItem.Description,
			&menuItem.Price,
			&menuItem.Category,
			&menuItem.DietaryType,
			&menuItem.SpiceLevel,
			&menuItem.Available,
			&menuItem.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		menuItems = append(menuItems, menuItem)
	}

	return menuItems, nil
}

func (db *DB) GetMenuItemByID(id int) (*models.MenuItem, error) {
	var menuItem models.MenuItem
	err := db.QueryRow(
		"SELECT id, restaurant_id, name, description, price, category, dietary_type, spice_level, available, created_at FROM menu_items WHERE id = $1",
		id,
	).Scan(
		&menuItem.ID,
		&menuItem.RestaurantID,
		&menuItem.Name,
		&menuItem.Description,
		&menuItem.Price,
		&menuItem.Category,
		&menuItem.DietaryType,
		&menuItem.SpiceLevel,
		&menuItem.Available,
		&menuItem.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("menu item with ID %d not found", id)
		}
		return nil, err
	}

	return &menuItem, nil
}

func (db *DB) CreateMenuItem(menuItem *models.MenuItem) error {
	query := `INSERT INTO menu_items (restaurant_id, name, description, price, category, dietary_type, spice_level, available) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at`
	err := db.QueryRow(
		query,
		menuItem.RestaurantID,
		menuItem.Name,
		menuItem.Description,
		menuItem.Price,
		menuItem.Category,
		menuItem.DietaryType,
		menuItem.SpiceLevel,
		menuItem.Available,
	).Scan(&menuItem.ID, &menuItem.CreatedAt)
	return err
}

// Order methods
func (db *DB) GetAllOrders() ([]models.Order, error) {
	rows, err := db.Query(`
		SELECT id, restaurant_id, customer_name, customer_phone, status, total_amount, tax_amount, discount, final_amount, payment_status, payment_method, billing_address, created_at, updated_at
		FROM orders ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.RestaurantID,
			&order.CustomerName,
			&order.CustomerPhone,
			&order.Status,
			&order.TotalAmount,
			&order.TaxAmount,
			&order.Discount,
			&order.FinalAmount,
			&order.PaymentStatus,
			&order.PaymentMethod,
			&order.BillingAddress,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		// Fetch associated order items
		orderItems, err := db.GetOrderItemsByOrderID(order.ID)
		if err != nil {
			return nil, err
		}
		order.OrderItems = orderItems
		
		orders = append(orders, order)
	}

	return orders, nil
}

func (db *DB) GetOrderByID(id int) (*models.Order, error) {
	var order models.Order
	err := db.QueryRow(`
		SELECT id, restaurant_id, customer_name, customer_phone, status, total_amount, tax_amount, discount, final_amount, payment_status, payment_method, billing_address, created_at, updated_at
		FROM orders WHERE id = $1
	`, id).Scan(
		&order.ID,
		&order.RestaurantID,
		&order.CustomerName,
		&order.CustomerPhone,
		&order.Status,
		&order.TotalAmount,
		&order.TaxAmount,
		&order.Discount,
		&order.FinalAmount,
		&order.PaymentStatus,
		&order.PaymentMethod,
		&order.BillingAddress,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order with ID %d not found", id)
		}
		return nil, err
	}

	// Fetch associated order items
	orderItems, err := db.GetOrderItemsByOrderID(order.ID)
	if err != nil {
		return nil, err
	}
	order.OrderItems = orderItems

	return &order, nil
}

func (db *DB) CreateOrder(order *models.Order) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var orderID int
	query := `
		INSERT INTO orders (
			restaurant_id, customer_name, customer_phone, status,
			total_amount, tax_amount, discount, final_amount,
			payment_status, payment_method, billing_address
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`
	err = tx.QueryRow(
		query,
		order.RestaurantID,
		order.CustomerName,
		order.CustomerPhone,
		order.Status,
		order.TotalAmount,
		order.TaxAmount,
		order.Discount,
		order.FinalAmount,
		order.PaymentStatus,
		order.PaymentMethod,
		order.BillingAddress,
	).Scan(&orderID)

	if err != nil {
		return err
	}

	// Assign the generated ID back to the order
	order.ID = orderID

	// Insert order items
	for _, item := range order.OrderItems {
		itemQuery := `
			INSERT INTO order_items (order_id, menu_item_id, quantity, price, notes)
			VALUES ($1, $2, $3, $4, $5)
		`
		_, err = tx.Exec(itemQuery, orderID, item.MenuItemID, item.Quantity, item.Price, item.Notes)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) GetOrderItemsByOrderID(orderID int) ([]models.OrderItem, error) {
	rows, err := db.Query(`
		SELECT oi.id, oi.order_id, oi.menu_item_id, mi.id, mi.restaurant_id, mi.name, mi.description, mi.price, mi.category, mi.dietary_type, mi.spice_level, mi.available, oi.quantity, oi.price, oi.notes, oi.subtotal
		FROM order_items oi
		JOIN menu_items mi ON oi.menu_item_id = mi.id
		WHERE oi.order_id = $1
		ORDER BY oi.id
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orderItems []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		var menuItem models.MenuItem

		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.MenuItemID,
			&menuItem.ID,
			&menuItem.RestaurantID,
			&menuItem.Name,
			&menuItem.Description,
			&menuItem.Price,
			&menuItem.Category,
			&menuItem.DietaryType,
			&menuItem.SpiceLevel,
			&menuItem.Available,
			&item.Quantity,
			&item.Price,
			&item.Notes,
			&item.Subtotal,
		)
		if err != nil {
			return nil, err
		}

		item.MenuItem = menuItem
		orderItems = append(orderItems, item)
	}

	return orderItems, nil
}