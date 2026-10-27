-- Migration: Create deals table
CREATE TABLE IF NOT EXISTS deals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    amount NUMERIC(15,2),
    account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    stage VARCHAR(100) NOT NULL DEFAULT 'Qualification' CHECK (stage IN ('Qualification','Needs Analysis','Proposal','Negotiation','Closed Won','Closed Lost')),
    probability INTEGER DEFAULT 10 CHECK (probability >= 0 AND probability <= 100),
    expected_close_date DATE,
    lead_source VARCHAR(100),
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_deals_stage ON deals(stage) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_deals_account ON deals(account_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_deals_contact ON deals(contact_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_deals_owner ON deals(owner_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_deals_created ON deals(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_deals_close_date ON deals(expected_close_date) WHERE deleted_at IS NULL;
