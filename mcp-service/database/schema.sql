-- Complete Database Schema for MCP Service with OAuth
-- This includes OAuth tables and Restaurant tables

-- ============================================
-- OAuth Tables
-- ============================================

-- User Profiles (OAuth users)
CREATE TABLE IF NOT EXISTS user_profiles (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    picture TEXT,
    provider VARCHAR(50),
    provider_user_id VARCHAR(255),
    status VARCHAR(20) DEFAULT 'active',
    role VARCHAR(20) DEFAULT 'user',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_login_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT unique_provider_user UNIQUE(provider, provider_user_id)
);

-- OAuth Clients (Dynamic Client Registration)
CREATE TABLE IF NOT EXISTS oauth_clients (
    id SERIAL PRIMARY KEY,
    client_id VARCHAR(255) UNIQUE NOT NULL,
    client_secret VARCHAR(255) NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    client_uri TEXT,
    logo_uri TEXT,
    redirect_uris JSONB NOT NULL DEFAULT '[]'::jsonb,
    grant_types JSONB NOT NULL DEFAULT '["authorization_code"]'::jsonb,
    response_types JSONB NOT NULL DEFAULT '["code"]'::jsonb,
    scope TEXT DEFAULT 'openid profile email',
    application_type VARCHAR(50) DEFAULT 'web',
    token_endpoint_auth_method VARCHAR(50) DEFAULT 'none',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    client_secret_expires_at BIGINT DEFAULT 0,
    active BOOLEAN DEFAULT true
);

-- OAuth Tokens (for revocation tracking)
CREATE TABLE IF NOT EXISTS oauth_tokens (
    id SERIAL PRIMARY KEY,
    token_id VARCHAR(255) UNIQUE NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    token_type VARCHAR(50) NOT NULL,
    scope TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    active BOOLEAN DEFAULT true,
    
    FOREIGN KEY (client_id) REFERENCES oauth_clients(client_id) ON DELETE CASCADE
);

-- ============================================
-- Restaurant Tables
-- ============================================

-- Restaurants
CREATE TABLE IF NOT EXISTS restaurants (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    phone_number TEXT,
    cuisine_type TEXT DEFAULT 'Indian',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Menu Items
CREATE TABLE IF NOT EXISTS menu_items (
    id SERIAL PRIMARY KEY,
    restaurant_id INTEGER REFERENCES restaurants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    category TEXT,
    dietary_type TEXT,
    spice_level TEXT,
    available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Orders
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
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Order Items
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id INTEGER REFERENCES menu_items(id),
    quantity INTEGER NOT NULL DEFAULT 1,
    price DECIMAL(10, 2) NOT NULL,
    notes TEXT,
    subtotal DECIMAL(10, 2) GENERATED ALWAYS AS (quantity * price) STORED
);

-- ============================================
-- Indexes for Performance
-- ============================================

-- OAuth indexes
CREATE INDEX IF NOT EXISTS idx_user_email ON user_profiles(email);
CREATE INDEX IF NOT EXISTS idx_user_status ON user_profiles(status);
CREATE INDEX IF NOT EXISTS idx_user_role ON user_profiles(role);
CREATE INDEX IF NOT EXISTS idx_user_provider ON user_profiles(provider, provider_user_id);

CREATE INDEX IF NOT EXISTS idx_oauth_clients_client_id ON oauth_clients(client_id);
CREATE INDEX IF NOT EXISTS idx_oauth_clients_active ON oauth_clients(active);

CREATE INDEX IF NOT EXISTS idx_oauth_tokens_token_id ON oauth_tokens(token_id);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_client_user ON oauth_tokens(client_id, user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_expires ON oauth_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_active ON oauth_tokens(active);

-- Restaurant indexes
CREATE INDEX IF NOT EXISTS idx_menu_items_restaurant ON menu_items(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_orders_restaurant ON orders(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);

-- ============================================
-- Triggers
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers
DROP TRIGGER IF EXISTS update_oauth_clients_updated_at ON oauth_clients;
CREATE TRIGGER update_oauth_clients_updated_at
    BEFORE UPDATE ON oauth_clients
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON user_profiles;
CREATE TRIGGER update_user_profiles_updated_at
    BEFORE UPDATE ON user_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
CREATE TRIGGER update_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Seed Default Admin User
-- ============================================

-- Insert default admin (will be skipped if already exists due to unique constraint)
INSERT INTO user_profiles (user_id, email, name, status, role, created_at) 
VALUES (
    'admin-default-vishal',
    'vishalkapadi17@hotmail.com',
    'Vishal Kapadi',
    'active',
    'admin',
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- ============================================
-- Sample Restaurant Data (Optional)
-- ============================================

-- Insert sample restaurants
INSERT INTO restaurants (name, address, phone_number, cuisine_type) VALUES
    ('Taj Mahal Restaurant', 'Connaught Place, New Delhi', '+91-11-12345678', 'Indian'),
    ('Surya Mahal', 'Linking Road, Mumbai', '+91-22-87654321', 'Indian'),
    ('Hyderabad House', 'Banjara Hills, Hyderabad', '+91-40-23456789', 'Indian')
ON CONFLICT DO NOTHING;

-- Note: Menu items will be inserted via application seed logic
