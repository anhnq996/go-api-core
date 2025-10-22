# Build stage
FROM golang:1.23-alpine AS builder

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

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=Asia/Ho_Chi_Minh

# Create app user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/apicore .

# Copy necessary files
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/database/migrations ./database/migrations

# Create directories
RUN mkdir -p storages/logs && \
    chown -R app:app /app

# Switch to app user
USER app

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/ping || exit 1

# Run application
CMD ["./apicore"]

