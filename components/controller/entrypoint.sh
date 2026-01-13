#!/bin/bash
set -e

echo "=== IDEKube Controller Entrypoint ==="
echo "Starting database migration process..."

# Run database migrations
echo "Running migrations..."
/app/migrate

# Check if migration was successful
if [ $? -ne 0 ]; then
    echo "ERROR: Database migration failed!"
    exit 1
fi

echo "Migration completed successfully!"
echo ""
echo "=== Starting IDEKube Controller ==="

# Start the controller
exec /app/idekube-controller
