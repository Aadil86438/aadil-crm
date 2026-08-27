-- Migration: Create notes table
CREATE TABLE IF NOT EXISTS notes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255),
    body TEXT NOT NULL,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    related_lead_id UUID REFERENCES leads(id) ON DELETE SET NULL,
    related_contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    related_account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    related_deal_id UUID REFERENCES deals(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_notes_owner ON notes(owner_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_notes_lead ON notes(related_lead_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_notes_contact ON notes(related_contact_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_notes_deal ON notes(related_deal_id) WHERE deleted_at IS NULL;
