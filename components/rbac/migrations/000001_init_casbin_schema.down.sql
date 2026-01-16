-- Drop OPA tables and related objects

-- Drop role bindings indexes
DROP INDEX IF EXISTS idx_opa_role_bindings_unique;
DROP INDEX IF EXISTS idx_opa_role_bindings_role;
DROP INDEX IF EXISTS idx_opa_role_bindings_subject;

-- Drop policies indexes
DROP INDEX IF EXISTS idx_opa_policies_action;
DROP INDEX IF EXISTS idx_opa_policies_object;
DROP INDEX IF EXISTS idx_opa_policies_subject;

-- Drop tables
DROP TABLE IF EXISTS opa_role_bindings;
DROP TABLE IF EXISTS opa_policies;
