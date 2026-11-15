include .env
export

export UID := $(shell id -u)
export GID := $(shell id -g)
export PROJECT_ROOT := $(shell pwd)

env-up:
	@docker compose up -d sso-postgres

env-down:
	@docker compose down --remove-orphans sso-postgres port-forwarder

env-cleanup:
	@read -p "Are you sure you want to remove the database volume? (y/n)" confrim; \
	if [ "$$confrim" = "y" ] || ["$$confrim" = "Y"]; then \
		docker compose down -v; \
		docker run --rm v $(PROJECT_ROOT):/app alpine rm -rf /app/.out/pgdata; \
		echo "Database volume removed."; \
	else \
		echo "Operation canceled."; \
	fi

env-port-forwarder:
	@docker compose up -d port-forwarder

env-port-close:
	@docker compose down port-forwarder

migrate-create:
	@if [ -z "$(seq)"]; then \
		echo "Error: Please provide a migration name using 'make migrate-create seq=<migration_name>'"; \
		exit 1; \
	fi; \
	docker compose run --rm todoapp-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up:
	@make migrare-action action=up

migrate-down:
	@make migrare-action action=down

migrate-action:
	@if [ -z "$(action)"]; then \
		echo "Error: Please provide a migration using 'make migrate-action action=<up|down>'"
		exit 1; \


