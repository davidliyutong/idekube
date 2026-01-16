-- Initialize OPA RBAC database schema
-- This script creates the necessary tables for OPA policy storage

-- Create OPA policy table
CREATE TABLE IF NOT EXISTS opa_policies (
    id SERIAL PRIMARY KEY,
    subject VARCHAR(255) NOT NULL,
    object VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_opa_policies_subject ON opa_policies(subject);
CREATE INDEX IF NOT EXISTS idx_opa_policies_object ON opa_policies(object);
CREATE INDEX IF NOT EXISTS idx_opa_policies_action ON opa_policies(action);

-- Create OPA role bindings table
CREATE TABLE IF NOT EXISTS opa_role_bindings (
    id SERIAL PRIMARY KEY,
    subject VARCHAR(255) NOT NULL,
    role VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_opa_role_bindings_subject ON opa_role_bindings(subject);
CREATE INDEX IF NOT EXISTS idx_opa_role_bindings_role ON opa_role_bindings(role);

-- Ensure unique constraint
CREATE UNIQUE INDEX IF NOT EXISTS idx_opa_role_bindings_unique ON opa_role_bindings(subject, role);

-- Add comments for documentation
COMMENT ON TABLE opa_policies IS 'Stores OPA RBAC policy rules';
COMMENT ON COLUMN opa_policies.subject IS 'Subject (user/role) for policy rules';
COMMENT ON COLUMN opa_policies.object IS 'Object (resource) for policy rules';
COMMENT ON COLUMN opa_policies.action IS 'Action (read/write/delete) for policy rules';

COMMENT ON TABLE opa_role_bindings IS 'Stores OPA role assignments to users';
COMMENT ON COLUMN opa_role_bindings.subject IS 'Subject (user) identifier';
COMMENT ON COLUMN opa_role_bindings.role IS 'Role identifier';
