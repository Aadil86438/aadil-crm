-- Migration: Add approval_status, crm_code, and company_name to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS crm_code VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS approval_status VARCHAR(50) DEFAULT 'pending';
ALTER TABLE users ADD COLUMN IF NOT EXISTS company_name VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_users_crm_code ON users(crm_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_approval_status ON users(approval_status) WHERE deleted_at IS NULL;

-- Ensure existing active users (initial admin/system users) remain approved and have CRMU01
UPDATE users 
SET approval_status = 'approved',
    crm_code = COALESCE(crm_code, 'CRMU01')
WHERE status = 'active' AND (approval_status IS NULL OR approval_status = 'pending');
