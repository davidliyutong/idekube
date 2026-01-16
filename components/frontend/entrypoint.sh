#!/bin/sh
set -e

# Set default values if not provided
export BACKEND_HOST=${BACKEND_HOST:-localhost}
export BACKEND_PORT=${BACKEND_PORT:-3000}
export RESOLVER=${RESOLVER:-127.0.0.11}

echo "Configuring nginx with backend: ${BACKEND_HOST}:${BACKEND_PORT}"
echo "DNS resolver: ${RESOLVER}"

# Substitute environment variables in nginx config template
envsubst '${BACKEND_HOST} ${BACKEND_PORT} ${RESOLVER}' < /etc/nginx/conf.d/default.conf.template > /etc/nginx/conf.d/default.conf

# Test nginx configuration
nginx -t

echo "Starting nginx..."
# Start nginx
exec nginx -g 'daemon off;'
