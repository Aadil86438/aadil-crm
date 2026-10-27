-- Seed default development admin user
-- Credentials:
-- Email: admin@crm.local
-- Password: admin123
-- CRM Access Code: CRMU01
INSERT INTO users (
    id,
    name,
    email,
    password_hash,
    role,
    status,
    approval_status,
    crm_code,
    company_name
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Local Admin',
    'admin@crm.local',
    '$2a$10$VgLQ2bERnIHYhgG8Ml6qR.WpkjjRfAhK97wOhfKln0dAx0lmWCiAG',
    'admin',
    'active',
    'approved',
    'CRMU01',
    'Local Dev Corp'
) ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    approval_status = 'approved',
    crm_code = 'CRMU01',
    status = 'active';
