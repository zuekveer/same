build:
	@docker compose build

rebuild:
	@docker compose down && docker compose up -d --build

run:
	@docker compose up -d

stop:
	@docker compose down

restart:
	@docker compose down && docker compose up -d

app:
	@docker compose exec app sh

clean:
	@docker compose down --rmi all -v

migrate-add: ## Create new migration file, usage: make migrate-add name=<migration_name>
	@go run github.com/pressly/goose/v3/cmd/goose@latest create $(name) sql -dir internal/database/migrations

lint:
	@golangci-lint run
