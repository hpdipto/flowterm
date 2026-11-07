DB_NAME ?= $(HOME)/.config/flowterm/flowterm.db

migration-create:
	goose -dir db/migrations create $(name) sql

migration-up:
	@mkdir -p $(dir $(DB_NAME))
	goose -dir db/migrations sqlite3 $(DB_NAME) up

migration-down:
	@mkdir -p $(dir $(DB_NAME))
	goose -dir db/migrations sqlite3 $(DB_NAME) down

migration-status:
	@mkdir -p $(dir $(DB_NAME))
	goose -dir db/migrations sqlite3 $(DB_NAME) status
