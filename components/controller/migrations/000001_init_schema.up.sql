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
    role VARCHAR(50) NOT NULL DEFAULT 'user', -- super_admin, admin, user
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
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(owner_type, owner_id, name)
);

CREATE INDEX idx_volumes_owner ON volumes(owner_type, owner_id);
CREATE INDEX idx_volumes_status ON volumes(status);
CREATE INDEX idx_volumes_deleted_at ON volumes(deleted_at
CREATE INDEX idx_volumes_owner ON volumes(owner_type, owner_id);
CREATE INDEX idx_volumes_status ON volumes(status);

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
CREATE INDEX idx_workspaces_template_id ON workspaces(template_id);
CREATE INDEX idx_workspaces_created_by ON workspaces(created_by);

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

-- Note: Admin user will be created automatically by the application
-- if ADMIN_PASSWORD environment variable is set
