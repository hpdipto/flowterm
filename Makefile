include .env

migration-create:
	goose -dir db/migrations create $(name) sql

migration-up:
	goose -dir db/migrations sqlite3 $(DB_NAME) up

migration-down:
	goose -dir db/migrations sqlite3 $(DB_NAME) down

migration-status:
	goose -dir db/migrations sqlite3 $(DB_NAME) status


