-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
    stock INTEGER NOT NULL CHECK (stock >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);

INSERT INTO products (id, name, description, price, stock) VALUES
-- Use UUIDs for product IDs
INSERT INTO products (id, name, description, price, stock) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Laptop', 'High-performance laptop', 999.99, 10),
    ('22222222-2222-2222-2222-222222222222', 'Mouse', 'Wireless mouse', 29.99, 50),
    ('33333333-3333-3333-3333-333333333333', 'Keyboard', 'Mechanical keyboard', 129.99, 25),
    ('44444444-4444-4444-4444-444444444444', 'Monitor', '24-inch monitor', 299.99, 15),
    ('55555555-5555-5555-5555-555555555555', 'Headphones', 'Noise-canceling headphones', 199.99, 20);