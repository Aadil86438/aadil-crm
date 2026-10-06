-- Migration: Create activities table
CREATE TABLE IF NOT EXISTS activities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL CHECK (type IN ('call','meeting','task','note','email')),
    subject VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(100),
    due_date TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    related_lead_id UUID REFERENCES leads(id) ON DELETE SET NULL,
    related_contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    related_account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    related_deal_id UUID REFERENCES deals(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_activities_type ON activities(type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_activities_owner ON activities(owner_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_activities_lead ON activities(related_lead_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_activities_contact ON activities(related_contact_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_activities_deal ON activities(related_deal_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_activities_created ON activities(created_at DESC) WHERE deleted_at IS NULL;
