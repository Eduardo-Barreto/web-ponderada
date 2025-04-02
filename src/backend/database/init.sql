-- src/database/init.sql

-- Enable UUID generation if needed (though we use Go's UUID for filenames)
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users Table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    profile_pic VARCHAR(255) DEFAULT '', -- Stores relative path like 'users/uuid.jpg'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Products Table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    value NUMERIC(10, 2) NOT NULL CHECK (value > 0), -- Example: Up to 99,999,999.99
    quantity INTEGER NOT NULL CHECK (quantity >= 0),
    image VARCHAR(255) DEFAULT '', -- Stores relative path like 'products/uuid.png'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Optional: Indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_products_description ON products(description);

-- TODO: Add trigger function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply the trigger to users table
DROP TRIGGER IF EXISTS set_timestamp_users ON users; -- Drop existing trigger if necessary
CREATE TRIGGER set_timestamp_users
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- Apply the trigger to products table
DROP TRIGGER IF EXISTS set_timestamp_products ON products; -- Drop existing trigger if necessary
CREATE TRIGGER set_timestamp_products
BEFORE UPDATE ON products
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();
