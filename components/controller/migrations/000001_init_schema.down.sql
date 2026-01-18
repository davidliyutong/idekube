-- Drop all tables in reverse order to respect foreign key constraints

DROP TABLE IF EXISTS workspace_transfers;
DROP TABLE IF EXISTS oauth_sessions;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS oidc_providers;
DROP TABLE IF EXISTS workspace_volumes;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS volumes;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS quotas;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS settings;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop extension (optional, may be used by other databases)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
