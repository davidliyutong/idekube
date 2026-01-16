# Database Migrations

This directory contains versioned database migration files for the IDEKube RBAC service.

## Migration Tool

We use [golang-migrate](https://github.com/golang-migrate/migrate) for database schema versioning and migration management.

## Directory Structure

```
migrations/
├── README.md                                    # This file
├── 000001_init_casbin_schema.up.sql           # Initial Casbin schema creation
└── 000001_init_casbin_schema.down.sql         # Initial Casbin schema rollback
```

## Migration Naming Convention

Migration files follow this naming pattern:
```
{version}_{description}.{direction}.sql
```

- **version**: 6-digit number (e.g., 000001, 000002)
- **description**: Snake_case description of the migration
- **direction**: `up` (apply) or `down` (rollback)

## How to Use

### Manual Migration (CLI Tool)

Use the built-in migrate command:

```bash
# Run all pending migrations
./bin/idekube-rbac-migrate -action up

# Rollback last migration
./bin/idekube-rbac-migrate -action down

# Check current migration version
./bin/idekube-rbac-migrate -action version

# Specify custom migrations path
./bin/idekube-rbac-migrate -action up -path ./migrations
```

Or use the Makefile:

```bash
# Run migrations up
make migrate-up

# Rollback last migration
make migrate-down

# Check migration version
make migrate-version
```

### Using golang-migrate CLI

Install golang-migrate:
```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Windows (using Scoop)
scoop install migrate
```

Run migrations:
```bash
# Apply all migrations
migrate -path migrations -database "postgresql://user:password@localhost:5432/idekube_rbac?sslmode=disable" up

# Rollback last migration
migrate -path migrations -database "postgresql://user:password@localhost:5432/idekube_rbac?sslmode=disable" down 1

# Check version
migrate -path migrations -database "postgresql://user:password@localhost:5432/idekube_rbac?sslmode=disable" version

# Force to a specific version (use with caution)
migrate -path migrations -database "postgresql://user:password@localhost:5432/idekube_rbac?sslmode=disable" force 1
```

## Database Schema

### casbin_rule Table

Stores Casbin RBAC policies and role assignments.

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL | Primary key |
| ptype | VARCHAR(100) | Policy type: 'p' (policy) or 'g' (grouping/role) |
| v0 | VARCHAR(100) | Subject (user/role) for policy rules |
| v1 | VARCHAR(100) | Object (resource) for policy rules, or role for grouping |
| v2 | VARCHAR(100) | Action (read/write/delete) for policy rules |
| v3 | VARCHAR(100) | Additional field for custom matchers |
| v4 | VARCHAR(100) | Additional field for custom matchers |
| v5 | VARCHAR(100) | Additional field for custom matchers |
| created_at | TIMESTAMP WITH TIME ZONE | Record creation timestamp |

### Indexes

- `idx_casbin_rule_ptype`: Index on policy type for quick filtering
- `idx_casbin_rule_v0`: Index on subject for user-based queries
- `idx_casbin_rule_v1`: Index on object for resource-based queries
- `idx_casbin_rule_v2`: Index on action for permission checks
- `idx_casbin_rule_ptype_v0_v1`: Composite index for common permission check patterns

## Creating New Migrations

To create a new migration:

1. Create two files with the next version number:
   - `{version}_{description}.up.sql` - Apply changes
   - `{version}_{description}.down.sql` - Revert changes

2. Example:
   ```bash
   touch migrations/000002_add_audit_fields.up.sql
   touch migrations/000002_add_audit_fields.down.sql
   ```

3. Write SQL for both up and down migrations

4. Test your migration:
   ```bash
   # Apply the migration
   make migrate-up
   
   # Test rollback
   make migrate-down
   
   # Re-apply
   make migrate-up
   ```

## Best Practices

1. **Always create both up and down migrations** - Ensure rollback capability
2. **Test migrations thoroughly** - Apply and rollback multiple times
3. **Keep migrations small and focused** - One logical change per migration
4. **Never modify existing migrations** - Once deployed, create a new migration instead
5. **Use transactions carefully** - Some DDL statements in PostgreSQL are transactional
6. **Document complex migrations** - Add comments explaining non-obvious changes

## Policy Seeding

The RBAC service can seed initial policies from the `configs/policy.csv` file. If the `casbin_rule` table is empty on startup, policies from the CSV file will be automatically loaded.

## Environment Setup

Make sure your PostgreSQL database is configured correctly in your `.env` file:

```env
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_password
POSTGRES_DB=idekube_rbac
```

## Troubleshooting

### Migration is marked as "dirty"

This happens when a migration fails partway through. To fix:

```bash
# Check current version
make migrate-version

# Force to the last known good version
migrate -path migrations -database "postgresql://..." force {version}

# Then try running migrations again
make migrate-up
```

### Connection refused

Ensure PostgreSQL is running and the connection parameters are correct:

```bash
# Test database connection
psql -h localhost -U postgres -d idekube_rbac -c "SELECT 1"
```

### Permission denied

The database user needs appropriate permissions:

```sql
-- Grant necessary permissions
GRANT ALL PRIVILEGES ON DATABASE idekube_rbac TO your_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO your_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO your_user;
```
