# Casbin Configuration Files

This directory contains Casbin RBAC configuration files.

## File Descriptions

### model.conf

Casbin model definition file that defines the structure of RBAC rules:

- `[request_definition]`: Defines request format (subject, object, action)
- `[policy_definition]`: Defines policy format
- `[role_definition]`: Defines role inheritance relationships
  - `g`: User-role relationship
  - `g2`: Resource-resource group relationship (optional)
- `[policy_effect]`: Defines policy effect (allow or deny)
- `[matchers]`: Defines matching rules
  - Supports wildcard `*`
  - Supports `keyMatch2` for path matching

**Note:** This file typically does not need to be modified unless you want to change the basic structure of the permission model.

### policy.csv

Casbin policy file that defines specific permission rules and role assignments:

Format:
```csv
# Policy rules
p, subject, object, action

# Role assignment
g, user, role
```

Example:
```csv
# Admin role can perform all operations on workspace
p, role:admin, workspace, read
p, role:admin, workspace, write
p, role:admin, workspace, delete

# User 1 is assigned as admin
g, user:1, role:admin
```

## Usage

### Database Mode

By default, the service loads policies from the PostgreSQL database. The `policy.csv` file is only used to load initial policies.

### File Mode

If you need to always load policies from file (e.g., in test environment), you can:

1. Modify `policy.csv`
2. Clear the policy table in the database
3. Restart the service

## Modifying Policies

### Via API

Recommended to modify policies dynamically via API:

```bash
# Assign role to user
curl -X POST http://localhost:8080/api/v1/rbac/assign-role \
  -H "Content-Type: application/json" \
  -d '{"user_id": 123, "role": "admin"}'
```

### Via File

1. Edit `policy.csv`
2. Clear database policy table:
   ```sql
   DELETE FROM casbin_rule;
   ```
3. Restart service

## Predefined Roles

### admin (Administrator)
- Full access to all resources
- Can manage users and organizations

### editor (Editor)
- workspace: read, write, execute
- template: read, write
- volume: read, write

### viewer (Viewer)
- workspace: read
- template: read
- volume: read

### workspace-admin (Workspace Administrator)
- All workspace permissions
- Limited permissions on other resources

### template-admin (Template Administrator)
- All template permissions
- Limited permissions on other resources

## Resource Types

- `workspace`: Workspace
- `template`: Template
- `volume`: Persistent volume
- `organization`: Organization
- `user`: User

## Action Types

- `read`: Read/view
- `write`: Write/update
- `delete`: Delete
- `execute`: Execute
- `create`: Create
- `manage`: Manage (includes all operations)

## Example Scenarios

### Scenario 1: Assign Viewer Role to New User

Add to `policy.csv`:
```csv
g, user:100, role:viewer
```

Or via API:
```bash
curl -X POST http://localhost:8080/api/v1/rbac/assign-role \
  -H "Content-Type: application/json" \
  -d '{"user_id": 100, "role": "viewer"}'
```

### Scenario 2: Create Custom Role

Define new role in `policy.csv`:
```csv
# Define data-scientist role permissions
p, role:data-scientist, workspace, read
p, role:data-scientist, workspace, execute
p, role:data-scientist, template, read
p, role:data-scientist, volume, read
p, role:data-scientist, volume, write

# Assign role
g, user:200, role:data-scientist
```

### Scenario 3: Permissions for Specific Resource

Grant user permissions on specific workspace:
```csv
# User 201 can access workspace ws-001
p, user:201, workspace:ws-001, read
p, user:201, workspace:ws-001, write
```

## Best Practices

1. **Use Roles Instead of Direct Permission Assignment**: Managing permissions through roles is easier to maintain
2. **Follow Principle of Least Privilege**: Only grant necessary permissions
3. **Use Role Inheritance**: Leverage Casbin's role inheritance feature to simplify management
4. **Regular Audits**: Regularly review and update permission policies
5. **Test Policies**: Thoroughly test permission policies before applying in production

## Debugging

### Check if Policy is Effective

```bash
curl -X POST http://localhost:8080/api/v1/rbac/check \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "resource_type": "workspace",
    "action": "read"
  }'
```

### View Policies in Database

```sql
SELECT * FROM casbin_rule;
```

## Reference Resources

- [Casbin Official Documentation](https://casbin.org/docs/overview)
- [RBAC Model](https://casbin.org/docs/rbac)
- [Policy Syntax](https://casbin.org/docs/syntax-for-models)
