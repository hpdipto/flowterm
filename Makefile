DB := flowterm.db

migration-create:
	goose -dir db/migrations create $(name) sql

migration-up:
	goose -dir db/migrations sqlite3 $(DB) up

migration-down:
	goose -dir db/migrations sqlite3 $(DB) down

migration-status:
	goose -dir db/migrations sqlite3 $(DB) status


