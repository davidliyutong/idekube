````markdown
# Database Migrations

This directory contains versioned database migration files for the IDEKube Controller.

## Migration Tool

We use [golang-migrate](https://github.com/golang-migrate/migrate) for database schema versioning and migration management.

## Directory Structure

```
migrations/
├── README.md                                    # This file
├── 000001_init_schema.up.sql                   # Initial schema creation
├── 000001_init_schema.down.sql                 # Initial schema rollback
├── 000002_add_user_mfa_fields.up.sql          # Add MFA fields to users table
├── 000002_add_user_mfa_fields.down.sql        # Remove MFA fields
├── 000003_add_oidc_redirect_url.up.sql        # Add redirect_url to OIDC providers
├── 000003_add_oidc_redirect_url.down.sql      # Remove redirect_url field
├── 000004_create_oauth_sessions.up.sql        # Create oauth_sessions table
├── 000004_create_oauth_sessions.down.sql      # Drop oauth_sessions table
├── 000005_refactor_api_keys_pk.up.sql         # Refactor API keys primary key
├── 000005_refactor_api_keys_pk.down.sql       # Revert API keys primary key
├── 000006_refactor_sessions_pk.up.sql         # Refactor sessions primary key
└── 000006_refactor_sessions_pk.down.sql       # Revert sessions primary key
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

### Automatic Migration (Application Startup)

Migrations run automatically when the controller starts. The application will:
1. Connect to the database
2. Run all pending migrations
3. Log the migration status

### Manual Migration (CLI Tool)

For manual migration management, use the `migrate` CLI tool:

#### Build the CLI tool
```bash
make build-migrate
# or
go build -o bin/migrate ./cmd/migrate
```

#### Run migrations up (apply all pending migrations)
```bash
./bin/migrate -action=up
```

#### Roll back the last migration
```bash
./bin/migrate -action=down
```

#### Check current migration version
```bash
./bin/migrate -action=version
```

#### Specify custom migrations path
```bash
./bin/migrate -action=up -path=/path/to/migrations
```

### Using golang-migrate CLI

Install the CLI tool:
```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Windows
scoop install migrate
```

#### Apply all pending migrations
```bash
migrate -path ./migrations -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" up
```

#### Roll back the last migration
```bash
migrate -path ./migrations -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" down 1
```

#### Check version
```bash
migrate -path ./migrations -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" version
```

#### Force a specific version (use with caution)
```bash
migrate -path ./migrations -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" force 3
```

## Creating New Migrations

### Step 1: Create migration files

Create two files for each migration (up and down):

```bash
# Determine next version number
NEXT_VERSION=$(printf "%06d" $(($(ls -1 migrations/*.up.sql | wc -l) + 1)))

# Create migration files
touch migrations/${NEXT_VERSION}_your_migration_name.up.sql
touch migrations/${NEXT_VERSION}_your_migration_name.down.sql
```

### Step 2: Write migration SQL

**Up migration** (`*_up.sql`): Changes to apply
```sql
-- Add new column
ALTER TABLE users ADD COLUMN new_field VARCHAR(255);

-- Create index
CREATE INDEX idx_users_new_field ON users(new_field);
```

**Down migration** (`*_down.sql`): How to revert
```sql
-- Remove index
DROP INDEX IF EXISTS idx_users_new_field;

-- Remove column
ALTER TABLE users DROP COLUMN IF EXISTS new_field;
```

### Step 3: Test the migration

```bash
# Test up migration
./bin/migrate -action=up

# Verify the change
psql -U user -d dbname -c "\d users"

# Test down migration (rollback)
./bin/migrate -action=down

# Verify rollback
psql -U user -d dbname -c "\d users"

# Re-apply if successful
./bin/migrate -action=up
```

## Migration Best Practices

### 1. **Always provide both up and down migrations**
- Every up migration must have a corresponding down migration
- Down migrations should cleanly revert all changes

### 2. **Make migrations idempotent when possible**
```sql
-- Good: Can be run multiple times safely
CREATE TABLE IF NOT EXISTS users (...);
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN;

-- Bad: Will fail if already exists
CREATE TABLE users (...);
ALTER TABLE users ADD COLUMN email_verified BOOLEAN;
```

### 3. **Use transactions for data migrations**
```sql
BEGIN;

-- Multiple statements here
UPDATE users SET status = 'active' WHERE status IS NULL;
ALTER TABLE users ALTER COLUMN status SET NOT NULL;

COMMIT;
```

### 4. **Test migrations thoroughly**
- Test on a copy of production data
- Test both up and down migrations
- Verify data integrity after migration

### 5. **Keep migrations small and focused**
- One logical change per migration
- Makes troubleshooting easier
- Easier to revert if needed

### 6. **Never modify existing migrations**
- Once applied to production, migrations are immutable
- Create a new migration to fix issues
- Exception: Migrations not yet in production

### 7. **Document complex migrations**
Add comments explaining:
- Why the change is needed
- What the migration does
- Any manual steps required
- Expected duration for large datasets

### 8. **Handle data carefully**
```sql
-- Good: Preserve data during column rename
ALTER TABLE users RENAME COLUMN old_name TO new_name;

-- Better: Safe migration with new column
ALTER TABLE users ADD COLUMN new_name VARCHAR(255);
UPDATE users SET new_name = old_name;
-- After verification in production:
-- CREATE MIGRATION: DROP COLUMN old_name
```

## Migration States

### Normal State
- Migration version is tracked in `schema_migrations` table
- `dirty` flag is `false`

### Dirty State
- A migration failed partway through
- `dirty` flag is `true`
- **Manual intervention required**

#### Recovering from Dirty State

1. **Inspect the database**
```bash
psql -U user -d dbname
SELECT * FROM schema_migrations;
```

2. **Manually fix the issue**
- Review what the migration was trying to do
- Check what actually got applied
- Manually complete or revert the changes

3. **Force the version**
```bash
# If migration completed successfully despite error
migrate -path ./migrations -database "..." force <version>

# If migration needs to be rolled back
migrate -path ./migrations -database "..." force <previous_version>
```

## Troubleshooting

### Migration fails with "duplicate key" error
- The migration may have partially applied
- Check what already exists in the database
- Modify the migration to use `IF NOT EXISTS` or `IF EXISTS`

### Migration fails with "relation does not exist"
- Check migration order
- Ensure dependencies are in earlier migrations
- Verify the down migration properly cleaned up

### Can't rollback a migration
- Check if data would be lost
- Ensure down migration handles edge cases
- May need to create a new migration instead

### Migration takes too long
- Consider breaking into smaller migrations
- Use `CONCURRENTLY` for index creation (doesn't lock table)
```sql
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
```

## Migration History

| Version | Description | Date | Notes |
|---------|-------------|------|-------|
| 000001 | Initial schema | 2026-01-13 | Base tables for users, organizations, workspaces, etc. |
| 000002 | Add MFA fields | 2026-01-13 | Support for multi-factor authentication |
| 000003 | Add OIDC redirect URL | 2026-01-13 | Complete OIDC provider configuration |
| 000004 | Create OAuth sessions table | 2026-01-13 | Support for OAuth state management |
| 000005 | Refactor API keys PK | 2026-01-13 | Use key_hash as primary key for better performance |
| 000006 | Refactor sessions PK | 2026-01-13 | Use session_token as primary key |

## References

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL ALTER TABLE Documentation](https://www.postgresql.org/docs/current/sql-altertable.html)
- [Database Migration Best Practices](https://www.prisma.io/dataguide/types/relational/migration-strategies)

````