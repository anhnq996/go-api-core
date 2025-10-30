# Build stage
FROM golang:1.25.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o apicore cmd/app/main.go

# Build migrate binary (optional, for running migrations in container)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate cmd/migrate/main.go

# Runtime stage
FROM golang:1.25.3-alpine

# Install runtime dependencies (including make and wget for commands)
RUN apk --no-cache add ca-certificates tzdata make bash git wget

# Set timezone
ENV TZ=Asia/Ho_Chi_Minh

# Create app user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/apicore .

# Copy migrate binary from builder
COPY --from=builder /app/migrate .

# Copy necessary files for make commands
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/database/migrations ./database/migrations
COPY --from=builder /app/keys ./keys
COPY --from=builder /app/translations ./translations
COPY --from=builder /app/Makefile ./Makefile

# Copy and setup entrypoint script (before switching user)
COPY build/docker/entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Create scripts directory (for make commands that might need it)
RUN mkdir -p ./scripts

# Create directories
RUN mkdir -p storages/log storages/app && \
    chown -R app:app /app

# Switch to app user
USER app

# Expose port
EXPOSE 3000

# Health check
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD wget --no-verbose --tries=1 --spider http://localhost:3000/ping || exit 1

# Use entrypoint script (optional: set AUTO_MIGRATE=true and AUTO_SEED=true)
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

# Run application
CMD ["./apicore"]

