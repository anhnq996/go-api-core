# 🐳 Docker Guide

## Quick Start

### Development (Local)

```bash
# 1. Start dev services (PostgreSQL + Redis)
make dev

# 2. Run migrations and seeders
make setup

# 3. Run app locally
make run
```

### Production (Full Docker)

```bash
# 1. Build and start all services
make docker-build
make docker-up

# 2. Run migrations in container
docker exec apicore-api ./apicore -migrate

# 3. Run seeders (optional)
docker exec apicore-api ./apicore -seed

# 4. Check logs
make docker-logs
```

## Docker Files

### Dockerfile

Multi-stage build:

1. **Builder stage**: Build Go binary
2. **Runtime stage**: Minimal Alpine image

**Features:**

- ✅ Multi-stage build (small image ~20MB)
- ✅ Non-root user
- ✅ Health check
- ✅ Timezone support
- ✅ Static binary (no CGO)

### docker-compose.yml (Development)

Services:

- PostgreSQL
- Redis
- Loki (logs)
- Grafana (monitoring)

**Usage:**

```bash
docker-compose up -d postgres redis
```

### docker-compose.prod.yml (Production)

All services including API:

- API (ApiCore)
- PostgreSQL
- Redis
- Loki
- Grafana

**Usage:**

```bash
docker-compose -f docker-compose.prod.yml up -d
```

## Commands

### Development

```bash
# Start dev services
make dev
# or
docker-compose up -d postgres redis

# Stop dev services
make dev-down
# or
docker-compose down

# View logs
docker-compose logs -f postgres
docker-compose logs -f redis
```

### Production

```bash
# Build image
make docker-build
# or
docker build -t apicore:latest .

# Start all services
make docker-up
# or
docker-compose -f docker-compose.prod.yml up -d

# View logs
make docker-logs
# or
docker-compose -f docker-compose.prod.yml logs -f

# Stop all services
make docker-down
# or
docker-compose -f docker-compose.prod.yml down

# Stop and remove volumes
docker-compose -f docker-compose.prod.yml down -v
```

## Build

### Local Build

```bash
make build
# Output: bin/apicore
```

### Docker Build

```bash
# Build image
docker build -t apicore:latest .

# With build args
docker build \
  --build-arg GO_VERSION=1.23 \
  -t apicore:v1.0.0 .

# Check image size
docker images apicore
```

## Run

### Run Local Binary

```bash
# Build first
make build

# Run binary
./bin/apicore
```

### Run in Docker

```bash
# Run container
docker run -d \
  --name apicore \
  -p 3000:3000 \
  -e DB_HOST=postgres \
  -e REDIS_HOST=redis \
  --network apicore_monitoring \
  apicore:latest

# Check logs
docker logs -f apicore

# Stop container
docker stop apicore
docker rm apicore
```

## Environment Variables

### Required

```env
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=apicore

REDIS_HOST=redis
REDIS_PORT=6379
```

### Optional

```env
DB_SSLMODE=disable
REDIS_PASSWORD=
REDIS_DB=0
SERVER_PORT=3000
LOKI_URL=http://loki:3100
```

### Using .env file

```bash
# Create .env
cp env.example .env

# Edit values
vim .env

# Run with docker-compose
docker-compose -f docker-compose.prod.yml --env-file .env up -d
```

## Migrations in Docker

### Method 1: Exec in Container

```bash
# After container is running
docker exec apicore-api ./apicore -migrate
docker exec apicore-api ./apicore -seed
```

### Method 2: Before Container Start

```bash
# In Dockerfile or docker-compose
command: sh -c "./apicore -migrate && ./apicore"
```

### Method 3: Init Container

```yaml
# In docker-compose.prod.yml
services:
  migrate:
    image: apicore:latest
    command: ./apicore -migrate
    depends_on:
      - postgres
    restart: "no"
```

## Networking

### Networks

- `monitoring` - Internal network for all services

### Access Services

From host:

- API: http://localhost:3000
- PostgreSQL: localhost:5432
- Redis: localhost:6379
- Grafana: http://localhost:3001
- Loki: http://localhost:3100

From containers:

- PostgreSQL: `postgres:5432`
- Redis: `redis:6379`
- Loki: `loki:3100`

## Volumes

### Persistent Data

```yaml
volumes:
  redis-data: # Redis AOF persistence
  postgres-data: # PostgreSQL data
  loki-data: # Loki chunks
  grafana-data: # Grafana dashboards
```

### Backup Volumes

```bash
# Backup PostgreSQL
docker run --rm \
  -v apicore_postgres-data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/postgres-backup.tar.gz /data

# Restore
docker run --rm \
  -v apicore_postgres-data:/data \
  -v $(pwd):/backup \
  alpine tar xzf /backup/postgres-backup.tar.gz -C /
```

## Health Checks

All services have health checks:

```bash
# Check health
docker-compose ps

# Should show (healthy) for all services
```

## Troubleshooting

### Container won't start

```bash
# Check logs
docker logs apicore-api

# Common issues:
# - DB connection failed → check DB_HOST
# - Port already in use → change port mapping
```

### Can't connect to database

```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Check network
docker network inspect apicore_monitoring

# Test connection
docker exec apicore-api sh -c "wget -O- http://postgres:5432"
```

### Can't connect to Redis

```bash
# Check Redis
docker exec apicore-redis redis-cli ping

# From API container
docker exec apicore-api sh -c "nc -zv redis 6379"
```

### Migration failed

```bash
# Check migration status
docker exec apicore-api ./apicore -migrate-version

# Force version
docker exec apicore-api ./apicore -migrate-force -version 1

# Run again
docker exec apicore-api ./apicore -migrate
```

## Production Deployment

### 1. Build Image

```bash
docker build -t apicore:v1.0.0 .
```

### 2. Push to Registry

```bash
# Tag for registry
docker tag apicore:v1.0.0 registry.example.com/apicore:v1.0.0

# Push
docker push registry.example.com/apicore:v1.0.0
```

### 3. Deploy

```bash
# Pull and run
docker pull registry.example.com/apicore:v1.0.0
docker-compose -f docker-compose.prod.yml up -d
```

### 4. Update

```bash
# Pull new version
docker pull registry.example.com/apicore:v1.0.1

# Restart services
docker-compose -f docker-compose.prod.yml up -d --no-deps api

# Run migrations if needed
docker exec apicore-api ./apicore -migrate
```

## Best Practices

### 1. Use Multi-Stage Builds

```dockerfile
# ✅ Current Dockerfile uses multi-stage
FROM golang:1.23-alpine AS builder
# ... build ...
FROM alpine:latest
# ... runtime ...
```

### 2. Non-Root User

```dockerfile
# ✅ Current Dockerfile creates app user
USER app
```

### 3. Health Checks

```dockerfile
# ✅ All services have health checks
HEALTHCHECK --interval=30s CMD wget --spider http://localhost:3000/ping
```

### 4. Resource Limits

```yaml
# Add to docker-compose.prod.yml
services:
  api:
    deploy:
      resources:
        limits:
          cpus: "1"
          memory: 512M
        reservations:
          cpus: "0.5"
          memory: 256M
```

## Commands Cheat Sheet

```bash
# Development
make dev              # Start postgres + redis
make migrate          # Run migrations
make seed             # Run seeders
make run              # Run app

# Production
make docker-build     # Build image
make docker-up        # Start all services
make docker-down      # Stop all services
make docker-logs      # View logs

# Utilities
make clean            # Clean artifacts
make test             # Run tests
make wire             # Generate Wire code
```

## Image Size

```bash
# Check image size
docker images apicore

# Expected:
# apicore    latest    ~20-30MB
```

## Security

### Recommendations

1. **Use secrets for passwords**

```bash
docker secret create db_password ./db_password.txt
```

2. **Scan image for vulnerabilities**

```bash
docker scan apicore:latest
```

3. **Use specific versions**

```dockerfile
FROM golang:1.23.4-alpine AS builder
FROM alpine:3.19
```

4. **Don't expose ports publicly**

```yaml
# Use reverse proxy (nginx, traefik)
ports:
  - "127.0.0.1:3000:3000" # Only localhost
```

## Monitoring

### View Metrics

```bash
# API metrics
curl http://localhost:3000/ping

# Redis metrics
docker exec apicore-redis redis-cli INFO

# PostgreSQL metrics
docker exec apicore-postgres psql -U postgres -c "SELECT * FROM pg_stat_activity;"
```

### Grafana Dashboards

1. Open: http://localhost:3001
2. Login: admin / admin
3. Create dashboard for:
   - Request rate
   - Error rate
   - Response time
   - Cache hit rate

## Resources

- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)
- [docker-compose Reference](https://docs.docker.com/compose/compose-file/)
