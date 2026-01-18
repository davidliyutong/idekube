-- IDEKube Controller Database Schema
-- PostgreSQL 14+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- Users table
-- Uses embedded Base, Profile, and Security structures
-- ============================================================================
CREATE TABLE users (
    -- Base fields
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    labels JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive, suspended
    extra_info JSONB,
    -- Profile fields
    identifier VARCHAR(255) NOT NULL UNIQUE, -- username
    display_name VARCHAR(255),
    icon_url TEXT,
    description TEXT,
    -- Security fields
    password_hash VARCHAR(255) NOT NULL,
    mfa_enabled BOOLEAN NOT NULL DEFAULT false,
    mfa_secret TEXT,
    mfa_backup_codes TEXT[],
    last_login_at TIMESTAMP WITH TIME ZONE,
    -- User-specific fields
    email VARCHAR(255),
    email_verified BOOLEAN NOT NULL DEFAULT false,
    role VARCHAR(50) NOT NULL DEFAULT 'user' -- super_admin, admin, power_user, user
);

CREATE INDEX idx_users_identifier ON users(identifier);
CREATE UNIQUE INDEX idx_users_email_not_null ON users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_labels ON users USING GIN (labels);

-- ============================================================================
-- Organizations table
-- Uses embedded Base and Profile structures
-- ============================================================================
CREATE TABLE organizations (
    -- Base fields
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    labels JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive, suspended
    extra_info JSONB,
    -- Profile fields
    identifier VARCHAR(255) NOT NULL UNIQUE, -- organization name
    display_name VARCHAR(255),
    icon_url TEXT,
    description TEXT,
    -- Organization-specific fields
    owner_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    settings JSONB
);

CREATE INDEX idx_organizations_identifier ON organizations(identifier);
CREATE INDEX idx_organizations_owner_id ON organizations(owner_id);
CREATE INDEX idx_organizations_status ON organizations(status);
CREATE INDEX idx_organizations_deleted_at ON organizations(deleted_at);
CREATE INDEX idx_organizations_labels ON organizations USING GIN (labels);

-- ============================================================================
-- Organization members table
-- ============================================================================
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

-- ============================================================================
-- Quotas table
-- Only belongs to Organization (removed owner_type)
-- Uses embedded QuotaLimits structure
-- ============================================================================
CREATE TABLE quotas (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    -- QuotaLimits fields
    cpu_millicores INT NOT NULL DEFAULT 8000,
    memory_mb INT NOT NULL DEFAULT 16384,
    storage_mb INT NOT NULL DEFAULT 51200,
    gpu_count INT NOT NULL DEFAULT 0,
    max_workspaces INT NOT NULL DEFAULT 10,
    max_volumes INT NOT NULL DEFAULT 20,
    timeout_seconds INT NOT NULL DEFAULT 0, -- 0 = no timeout/unlimited; >0 = timeout in seconds; NULL not allowed
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_quotas_organization_id ON quotas(organization_id);

-- ============================================================================
-- Templates table
-- System-level concept (removed owner_type, owner_id, visibility)
-- Uses embedded Base, Profile, and QuotaLimits structures
-- ============================================================================
CREATE TABLE templates (
    -- Base fields
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    labels JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive, archived
    extra_info JSONB,
    -- Profile fields
    identifier VARCHAR(255) NOT NULL UNIQUE, -- template name
    display_name VARCHAR(255),
    icon_url TEXT,
    description TEXT,
    -- Template-specific fields
    image_ref VARCHAR(500) NOT NULL,
    template_yaml TEXT NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT false,
    -- Default QuotaLimits (with prefix default_)
    default_cpu_millicores INT NOT NULL DEFAULT 1000,
    default_memory_mb INT NOT NULL DEFAULT 2048,
    default_storage_mb INT NOT NULL DEFAULT 10240,
    default_gpu_count INT NOT NULL DEFAULT 0,
    default_max_workspaces INT NOT NULL DEFAULT 10,
    default_max_volumes INT NOT NULL DEFAULT 20,
    default_timeout_seconds INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_templates_identifier ON templates(identifier);
CREATE INDEX idx_templates_is_public ON templates(is_public);
CREATE INDEX idx_templates_status ON templates(status);
CREATE INDEX idx_templates_labels ON templates USING GIN (labels);

-- ============================================================================
-- Volumes table
-- All volumes belong to Organization (removed owner_type)
-- Uses embedded Base, Profile, and K8SResources structures
-- ============================================================================
CREATE TABLE volumes (
    -- Base fields
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    labels JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, bound, failed
    extra_info JSONB,
    -- Profile fields
    identifier VARCHAR(255) NOT NULL, -- volume name
    display_name VARCHAR(255),
    icon_url TEXT,
    description TEXT,
    -- Volume-specific fields
    size_mb INT NOT NULL,
    storage_class VARCHAR(255),
    access_mode VARCHAR(50) NOT NULL DEFAULT 'ReadWriteOnce',
    is_public BOOLEAN NOT NULL DEFAULT false,
    owner_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    -- K8SResources fields
    k8s_namespace VARCHAR(255),
    k8s_deployment_name VARCHAR(255),
    k8s_service_name VARCHAR(255),
    k8s_ingress_name VARCHAR(255),
    k8s_pvc_name VARCHAR(255),
    UNIQUE(owner_id, identifier)
);

CREATE INDEX idx_volumes_identifier ON volumes(identifier);
CREATE INDEX idx_volumes_owner_id ON volumes(owner_id);
CREATE INDEX idx_volumes_status ON volumes(status);
CREATE INDEX idx_volumes_deleted_at ON volumes(deleted_at);
CREATE INDEX idx_volumes_is_public ON volumes(is_public);
CREATE INDEX idx_volumes_labels ON volumes USING GIN (labels);

-- ============================================================================
-- Workspaces table
-- All workspaces belong to Organization (removed owner_type)
-- Uses embedded Base, Profile, K8SResources, and QuotaLimits structures
-- ============================================================================
CREATE TABLE workspaces (
    -- Base fields
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    labels JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, starting, running, stopped, failed
    extra_info JSONB,
    -- Profile fields
    identifier VARCHAR(255) NOT NULL, -- workspace name
    display_name VARCHAR(255),
    icon_url TEXT,
    description TEXT,
    -- Workspace-specific fields (immutable after creation)
    template_id BIGINT NOT NULL REFERENCES templates(id) ON DELETE RESTRICT,
    template_snapshot JSONB,
    parameters JSONB,
    -- K8SResources fields (managed by housekeeper)
    k8s_namespace VARCHAR(255),
    k8s_deployment_name VARCHAR(255),
    k8s_service_name VARCHAR(255),
    k8s_ingress_name VARCHAR(255),
    k8s_pvc_name VARCHAR(255),
    -- QuotaLimits fields
    cpu_millicores INT NOT NULL DEFAULT 1000,
    memory_mb INT NOT NULL DEFAULT 2048,
    storage_mb INT NOT NULL DEFAULT 10240,
    gpu_count INT NOT NULL DEFAULT 0,
    max_workspaces INT NOT NULL DEFAULT 10,
    max_volumes INT NOT NULL DEFAULT 20,
    timeout_seconds INT NOT NULL DEFAULT 0,
    -- Other workspace fields
    is_public BOOLEAN NOT NULL DEFAULT false,
    accessed_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    owner_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    target_status VARCHAR(50) NOT NULL DEFAULT 'running',
    UNIQUE(owner_id, identifier)
);

CREATE INDEX idx_workspaces_identifier ON workspaces(identifier);
CREATE INDEX idx_workspaces_owner_id ON workspaces(owner_id);
CREATE INDEX idx_workspaces_status ON workspaces(status);
CREATE INDEX idx_workspaces_template_id ON workspaces(template_id);
CREATE INDEX idx_workspaces_created_by ON workspaces(created_by);
CREATE INDEX idx_workspaces_deleted_at ON workspaces(deleted_at);
CREATE INDEX idx_workspaces_is_public ON workspaces(is_public);
CREATE INDEX idx_workspaces_labels ON workspaces USING GIN (labels);

-- ============================================================================
-- Workspace volumes table
-- ============================================================================
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

-- ============================================================================
-- OIDC providers table
-- ============================================================================
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

-- ============================================================================
-- Audit logs table
-- ============================================================================
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

-- ============================================================================
-- Sessions table
-- ============================================================================
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

-- ============================================================================
-- OAuth sessions table
-- ============================================================================
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

-- ============================================================================
-- Workspace transfers table
-- ============================================================================
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

-- ============================================================================
-- Settings table
-- ============================================================================
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

-- ============================================================================
-- Function and Triggers for updated_at
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

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

CREATE TRIGGER update_oauth_sessions_updated_at BEFORE UPDATE ON oauth_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_workspace_transfers_updated_at BEFORE UPDATE ON workspace_transfers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_settings_updated_at BEFORE UPDATE ON settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Comments
-- ============================================================================
COMMENT ON TABLE settings IS 'Stores system configuration settings';
COMMENT ON COLUMN settings.key IS 'Unique configuration key';
COMMENT ON COLUMN settings.value IS 'Configuration value as string';
COMMENT ON COLUMN settings.value_type IS 'Type of the value (string, int, bool)';
COMMENT ON COLUMN settings.is_public IS 'Whether non-admin users can read this setting';

-- ============================================================================
-- Initialize default settings
-- ============================================================================
INSERT INTO settings (key, value, value_type, description, category, is_public) VALUES
    ('access_token_expiration_minutes', '15', 'int', 'Access token expiration time in minutes', 'auth', false),
    ('refresh_token_expiration_days', '30', 'int', 'Refresh token expiration time in days', 'auth', false),
    ('login_max_attempts', '5', 'int', 'Maximum login attempts before account is temporarily locked', 'security', false),
    ('login_ban_duration_minutes', '15', 'int', 'Duration of temporary account lock in minutes', 'security', false);

-- Note: Admin user will be created automatically by the application
-- if ADMIN_PASSWORD environment variable is set
