-- IDEKube Controller Database Schema
-- PostgreSQL 14+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user', -- super_admin, admin, power_user, user
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive, suspended
    avatar_url TEXT,
    display_name VARCHAR(255),
    extra_info JSONB,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    mfa_enabled BOOLEAN NOT NULL DEFAULT false,
    mfa_secret TEXT,
    mfa_backup_codes TEXT[],
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_email_verified ON users(email_verified);
CREATE INDEX idx_users_mfa_enabled ON users(mfa_enabled);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Organizations table
CREATE TABLE organizations (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    name VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255),
    description TEXT,
    avatar_url TEXT,
    owner_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    settings JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_organizations_name ON organizations(name);
CREATE INDEX idx_organizations_owner_id ON organizations(owner_id);
CREATE INDEX idx_organizations_deleted_at ON organizations(deleted_at);

-- Organization members table
CREATE TABLE organization_members (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member', -- owner, admin, member
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, user_id)
);

CREATE INDEX idx_organization_members_org_id ON organization_members(organization_id);
CREATE INDEX idx_organization_members_user_id ON organization_members(user_id);

-- Quotas table
CREATE TABLE quotas (
    id BIGSERIAL PRIMARY KEY,
    owner_type VARCHAR(50) NOT NULL, -- user, organization
    owner_id BIGINT NOT NULL,
    cpu_millicores INT NOT NULL DEFAULT 8000,
    memory_mb INT NOT NULL DEFAULT 16384,
    storage_mb INT NOT NULL DEFAULT 51200,
    gpu_count INT NOT NULL DEFAULT 0,
    max_workspaces INT NOT NULL DEFAULT 10,
    max_volumes INT NOT NULL DEFAULT 20,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(owner_type, owner_id)
);

CREATE INDEX idx_quotas_owner ON quotas(owner_type, owner_id);

-- Templates table
CREATE TABLE templates (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    description TEXT,
    image_ref VARCHAR(500) NOT NULL,
    template_yaml TEXT NOT NULL,
    icon_url TEXT,
    is_public BOOLEAN NOT NULL DEFAULT false,
    owner_type VARCHAR(50),
    owner_id BIGINT,
    default_cpu_millicores INT NOT NULL DEFAULT 1000,
    default_memory_mb INT NOT NULL DEFAULT 2048,
    default_storage_mb INT NOT NULL DEFAULT 10240,
    labels JSONB,
    visibility VARCHAR(20) NOT NULL DEFAULT 'private', -- public, organization, private
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_templates_name ON templates(name);
CREATE INDEX idx_templates_is_public ON templates(is_public);
CREATE INDEX idx_templates_owner ON templates(owner_type, owner_id);
CREATE INDEX idx_templates_visibility ON templates(visibility);
CREATE INDEX idx_templates_labels ON templates USING GIN (labels);

-- Volumes table
CREATE TABLE volumes (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    description TEXT,
    owner_type VARCHAR(50) NOT NULL,
    owner_id BIGINT NOT NULL,
    size_mb INT NOT NULL,
    storage_class VARCHAR(255),
    access_mode VARCHAR(50) NOT NULL DEFAULT 'ReadWriteOnce',
    pvc_name VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    labels JSONB,
    organization_id BIGINT REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(owner_type, owner_id, name)
);

CREATE INDEX idx_volumes_owner ON volumes(owner_type, owner_id);
CREATE INDEX idx_volumes_status ON volumes(status);
CREATE INDEX idx_volumes_deleted_at ON volumes(deleted_at);
CREATE INDEX idx_volumes_organization_id ON volumes(organization_id);
CREATE INDEX idx_volumes_labels ON volumes USING GIN (labels);

-- Workspaces table
CREATE TABLE workspaces (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    description TEXT,
    owner_type VARCHAR(50) NOT NULL,
    owner_id BIGINT NOT NULL,
    template_id BIGINT NOT NULL REFERENCES templates(id) ON DELETE RESTRICT,
    cpu_millicores INT NOT NULL,
    memory_mb INT NOT NULL,
    storage_mb INT NOT NULL,
    current_status VARCHAR(50) NOT NULL DEFAULT 'stopped',
    target_status VARCHAR(50) NOT NULL DEFAULT 'running',
    k8s_namespace VARCHAR(255),
    k8s_deployment_name VARCHAR(255),
    k8s_service_name VARCHAR(255),
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    labels JSONB,
    organization_id BIGINT REFERENCES organizations(id) ON DELETE CASCADE,
    is_shared BOOLEAN NOT NULL DEFAULT false,
    accessed_at TIMESTAMP WITH TIME ZONE,
    timeout_seconds INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(owner_type, owner_id, name)
);

CREATE INDEX idx_workspaces_owner ON workspaces(owner_type, owner_id);
CREATE INDEX idx_workspaces_status ON workspaces(current_status);
CREATE INDEX idx_workspaces_template_id ON workspaces(template_id);
CREATE INDEX idx_workspaces_created_by ON workspaces(created_by);
CREATE INDEX idx_workspaces_deleted_at ON workspaces(deleted_at);
CREATE INDEX idx_workspaces_organization_id ON workspaces(organization_id);
CREATE INDEX idx_workspaces_is_shared ON workspaces(is_shared);
CREATE INDEX idx_workspaces_labels ON workspaces USING GIN (labels);

-- Workspace volumes table
CREATE TABLE workspace_volumes (
    id BIGSERIAL PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    volume_id BIGINT NOT NULL REFERENCES volumes(id) ON DELETE CASCADE,
    mount_path VARCHAR(500) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(workspace_id, volume_id),
    UNIQUE(workspace_id, mount_path)
);

CREATE INDEX idx_workspace_volumes_workspace_id ON workspace_volumes(workspace_id);
CREATE INDEX idx_workspace_volumes_volume_id ON workspace_volumes(volume_id);

-- OIDC providers table
CREATE TABLE oidc_providers (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255),
    issuer_url VARCHAR(500) NOT NULL,
    client_id VARCHAR(500) NOT NULL,
    client_secret VARCHAR(500) NOT NULL,
    scopes TEXT[],
    redirect_url VARCHAR(500) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_oidc_providers_enabled ON oidc_providers(enabled);

-- API keys table
CREATE TABLE api_keys (
    key_hash VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scopes TEXT[],
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_expires_at ON api_keys(expires_at);
CREATE INDEX idx_api_keys_revoked_at ON api_keys(revoked_at);

-- Webhooks table
CREATE TABLE webhooks (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(500) NOT NULL,
    secret VARCHAR(255),
    owner_type VARCHAR(50) NOT NULL,
    owner_id BIGINT NOT NULL,
    events TEXT[] NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhooks_owner ON webhooks(owner_type, owner_id);
CREATE INDEX idx_webhooks_enabled ON webhooks(enabled);

-- Audit logs table
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    username VARCHAR(255),
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- Sessions table
CREATE TABLE sessions (
    session_token VARCHAR(500) PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_activity_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_last_activity ON sessions(last_activity_at);

-- OAuth sessions table
CREATE TABLE oauth_sessions (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(500) NOT NULL UNIQUE,
    value TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_oauth_sessions_key ON oauth_sessions(key);
CREATE INDEX idx_oauth_sessions_expires_at ON oauth_sessions(expires_at);

-- Workspace transfers table
CREATE TABLE workspace_transfers (
    id BIGSERIAL PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    from_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_username VARCHAR(255) NOT NULL,
    to_user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, accepted, rejected, cancelled
    message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    responded_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_workspace_transfers_workspace_id ON workspace_transfers(workspace_id);
CREATE INDEX idx_workspace_transfers_from_user_id ON workspace_transfers(from_user_id);
CREATE INDEX idx_workspace_transfers_to_user_id ON workspace_transfers(to_user_id);
CREATE INDEX idx_workspace_transfers_status ON workspace_transfers(status);

-- Function to automatically update updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_quotas_updated_at BEFORE UPDATE ON quotas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_templates_updated_at BEFORE UPDATE ON templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_volumes_updated_at BEFORE UPDATE ON volumes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_workspaces_updated_at BEFORE UPDATE ON workspaces
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_oidc_providers_updated_at BEFORE UPDATE ON oidc_providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_webhooks_updated_at BEFORE UPDATE ON webhooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_oauth_sessions_updated_at BEFORE UPDATE ON oauth_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_workspace_transfers_updated_at BEFORE UPDATE ON workspace_transfers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- OPA RBAC Tables

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
-- Composite index for common queries
CREATE INDEX IF NOT EXISTS idx_opa_policies_subject_object ON opa_policies(subject, object);
CREATE INDEX IF NOT EXISTS idx_opa_policies_subject_object_action ON opa_policies(subject, object, action);

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

-- Add triggers for OPA tables
CREATE TRIGGER update_opa_policies_updated_at BEFORE UPDATE ON opa_policies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_opa_role_bindings_updated_at BEFORE UPDATE ON opa_role_bindings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE opa_policies IS 'Stores OPA RBAC policy rules';
COMMENT ON COLUMN opa_policies.subject IS 'Subject (user/role) for policy rules';
COMMENT ON COLUMN opa_policies.object IS 'Object (resource) for policy rules';
COMMENT ON COLUMN opa_policies.action IS 'Action (read/write/delete) for policy rules';

COMMENT ON TABLE opa_role_bindings IS 'Stores OPA role assignments to users';
COMMENT ON COLUMN opa_role_bindings.subject IS 'Subject (user) identifier';
COMMENT ON COLUMN opa_role_bindings.role IS 'Role identifier';

-- Settings table
CREATE TABLE settings (
    key VARCHAR(255) PRIMARY KEY,
    value VARCHAR(500) NOT NULL,
    value_type VARCHAR(50) NOT NULL DEFAULT 'string', -- string, int, bool
    description TEXT,
    category VARCHAR(100) NOT NULL DEFAULT 'general', -- general, auth, security, etc.
    is_public BOOLEAN NOT NULL DEFAULT false, -- if true, can be read without admin privileges
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_settings_category ON settings(category);
CREATE INDEX idx_settings_is_public ON settings(is_public);

-- Add trigger for settings
CREATE TRIGGER update_settings_updated_at BEFORE UPDATE ON settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE settings IS 'Stores system configuration settings';
COMMENT ON COLUMN settings.key IS 'Unique configuration key';
COMMENT ON COLUMN settings.value IS 'Configuration value as string';
COMMENT ON COLUMN settings.value_type IS 'Type of the value (string, int, bool)';
COMMENT ON COLUMN settings.is_public IS 'Whether non-admin users can read this setting';

-- Initialize default settings
INSERT INTO settings (key, value, value_type, description, category, is_public) VALUES
    ('access_token_expiration_minutes', '15', 'int', 'Access token expiration time in minutes', 'auth', false),
    ('refresh_token_expiration_days', '30', 'int', 'Refresh token expiration time in days', 'auth', false),
    ('login_max_attempts', '5', 'int', 'Maximum login attempts before account is temporarily locked', 'security', false),
    ('login_ban_duration_minutes', '15', 'int', 'Duration of temporary account lock in minutes', 'security', false);

-- Note: Admin user will be created automatically by the application
-- if ADMIN_PASSWORD environment variable is set
