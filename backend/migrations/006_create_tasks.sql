-- Migration: Create tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    subject VARCHAR(255) NOT NULL,
    description TEXT,
    due_date TIMESTAMP WITH TIME ZONE,
    priority VARCHAR(50) DEFAULT 'Medium' CHECK (priority IN ('Low','Medium','High')),
    status VARCHAR(100) DEFAULT 'Not Started' CHECK (status IN ('Not Started','In Progress','Completed','Deferred')),
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    related_lead_id UUID REFERENCES leads(id) ON DELETE SET NULL,
    related_contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    related_account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    related_deal_id UUID REFERENCES deals(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_tasks_owner ON tasks(owner_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_lead ON tasks(related_lead_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_contact ON tasks(related_contact_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_deal ON tasks(related_deal_id) WHERE deleted_at IS NULL;
