-- Migration: Create lead_conversions table
CREATE TABLE IF NOT EXISTS lead_conversions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lead_id UUID NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
    contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    deal_id UUID REFERENCES deals(id) ON DELETE SET NULL,
    converted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    converted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_lead_conversions_lead ON lead_conversions(lead_id);
CREATE INDEX IF NOT EXISTS idx_lead_conversions_contact ON lead_conversions(contact_id);
CREATE INDEX IF NOT EXISTS idx_lead_conversions_account ON lead_conversions(account_id);
CREATE INDEX IF NOT EXISTS idx_lead_conversions_deal ON lead_conversions(deal_id);
