.PHONY: build
build:
	@docker compose build

.PHONY: rebuild
rebuild:
	@docker compose down && docker compose up -d --build

.PHONY: run
run:
	@docker compose up -d

.PHONY: stop
stop:
	@docker compose down

.PHONY: restart
restart:
	@docker compose down && docker compose up -d

.PHONY: app
app:
	@docker compose exec app sh

.PHONY: clean
clean:
	@docker compose down --rmi all
