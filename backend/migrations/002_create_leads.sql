-- Migration: Create leads table
CREATE TABLE IF NOT EXISTS leads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    company VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(50),
    mobile VARCHAR(50),
    website VARCHAR(500),
    lead_source VARCHAR(100) CHECK (lead_source IN ('Website','Referral','Advertisement','Cold Call','Email','Social Media','Campaign','Other')),
    lead_status VARCHAR(100) NOT NULL DEFAULT 'New' CHECK (lead_status IN ('New','Contacted','Qualified','Unqualified','Converted')),
    industry VARCHAR(100),
    job_title VARCHAR(255),
    annual_revenue NUMERIC(15,2),
    num_employees INTEGER,
    rating VARCHAR(50),
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    description TEXT,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    is_converted BOOLEAN NOT NULL DEFAULT FALSE,
    converted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_leads_email ON leads(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_leads_status ON leads(lead_status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_leads_owner ON leads(owner_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_leads_created ON leads(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_leads_company ON leads(company) WHERE deleted_at IS NULL;
