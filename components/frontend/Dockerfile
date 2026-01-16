# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json yarn.lock ./

# Install dependencies
RUN yarn install --frozen-lockfile

# Copy source files
COPY . .

# Build application (skip type checking for faster builds)
RUN yarn build:prod

# Production stage
FROM nginx:alpine

# Install gettext for envsubst (already included in nginx:alpine)
# Copy nginx configuration template
COPY nginx.conf.template /etc/nginx/conf.d/default.conf.template

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Copy built files from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Set default environment variables
ENV BACKEND_HOST=localhost \
    BACKEND_PORT=3000

# Expose port 80
EXPOSE 80

# Health check using curl (available in nginx:alpine)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost/ || exit 1

# Use custom entrypoint
ENTRYPOINT ["/entrypoint.sh"]
