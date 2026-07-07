.PHONY: dev-api dev-frontend db-up db-down migrate build up down

# ─── Database ────────────────────────────────────────────
db-up:
	docker compose up -d postgres

db-down:
	docker compose down

# ─── API (Go) ───────────────────────────────────────────
dev-api:
	cd api && air run ./cmd/server

build-api:
	cd api && go build -o bin/server ./cmd/server

seed:
	cd api && go run ./cmd/seed

# ─── Frontend (Next.js) ─────────────────────────────────
dev-frontend:
	cd nextjs && npm run dev

build-frontend:
	cd nextjs && npm run build

# ─── Migrations ─────────────────────────────────────────
migrate:
	psql "$(DATABASE_URL)" -f api/migrations/001_initial.sql

# ─── Docker (full stack) ────────────────────────────────
build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

# ─── Help ───────────────────────────────────────────────
deploy-api:
	fly deploy

seed2:
	cd api && go run ./cmd/seed seed/temples2.json

help:
	@echo "Targets:"
	@echo "  make db-up          Start Postgres"
	@echo "  make db-down        Stop Postgres"
	@echo "  make dev-api        Run Go API (with air hot-reload)"
	@echo "  make dev-frontend   Run Next.js dev server"
	@echo "  make migrate        Run SQL migrations"
	@echo "  make up             Start full stack with Docker Compose"
	@echo "  make down           Stop Docker Compose"
