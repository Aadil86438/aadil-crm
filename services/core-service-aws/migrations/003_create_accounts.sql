-- Migration: Create accounts table (before contacts, as contacts reference accounts)
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    website VARCHAR(500),
    industry VARCHAR(100),
    phone VARCHAR(50),
    email VARCHAR(255),
    num_employees INTEGER,
    annual_revenue NUMERIC(15,2),
    account_type VARCHAR(100) DEFAULT 'Prospect' CHECK (account_type IN ('Customer','Prospect','Partner','Reseller','Other')),
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_accounts_name ON accounts(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_accounts_owner ON accounts(owner_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_accounts_type ON accounts(account_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_accounts_created ON accounts(created_at DESC) WHERE deleted_at IS NULL;
