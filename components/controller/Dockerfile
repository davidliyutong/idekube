# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build both the controller and migrate tool
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o idekube-controller ./cmd/controller
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate ./cmd/migrate

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates bash

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/idekube-controller .
COPY --from=builder /app/migrate .

# Copy migrations directory
COPY --from=builder /app/migrations ./migrations
# Copy configs directory
COPY --from=builder /app/configs ./configs

# Copy entrypoint script
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# Use the entrypoint script
ENTRYPOINT ["./entrypoint.sh"]
