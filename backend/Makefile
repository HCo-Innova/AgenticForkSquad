.PHONY: help deps migrate seed setup clean logs verify

help: ## Show this help
	@echo "📋 Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

deps: ## Install Go dependencies
	@echo "📦 Installing dependencies..."
	@docker compose exec backend go mod download
	@docker compose exec backend go mod tidy
	@echo "✅ Dependencies installed"

migrate: ## Run database migrations (Tiger Cloud via DATABASE_URL)
	@echo "🔄 Running migrations (Tiger)..."
	@docker compose exec -e DATABASE_URL="$(DATABASE_URL)" backend goose -dir ./migrations postgres "$$DATABASE_URL" up
	@echo "✅ Migrations completed"

seed: ## Seed database with test data (Tiger)
	@echo "🌱 Seeding database (10K orders) on Tiger..."
	@docker compose exec -e DATABASE_URL="$(DATABASE_URL)" -e PGPASSWORD="$(TIGER_DB_PASSWORD)" backend go run /app/cmd/api/seed/main.go
	@echo "✅ Database seeded"

setup: migrate seed ## Full database setup (migrate + seed)
	@echo "✅ Database ready!"

clean: ## Clean all data from database (Tiger)
	@echo "🧹 Cleaning database (Tiger)..."
	@docker compose exec -e PGPASSWORD="$(TIGER_DB_PASSWORD)" backend psql "$(DATABASE_URL)" -c "TRUNCATE users, orders, payments RESTART IDENTITY CASCADE"
	@echo "✅ Database cleaned"

logs: ## Show backend logs
	@docker compose logs -f backend

verify: ## Verify database content (Tiger)
	@echo "🔍 Verifying database (Tiger)..."
	@docker compose exec -e PGPASSWORD="$(TIGER_DB_PASSWORD)" backend psql "$(DATABASE_URL)" -c "SELECT 'users' as table_name, COUNT(*) as count FROM users UNION ALL SELECT 'orders', COUNT(*) FROM orders UNION ALL SELECT 'payments', COUNT(*) FROM payments;"