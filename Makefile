include .env
export

export UID := $(shell id -u)
export GID := $(shell id -g)
export PROJECT_ROOT := $(shell pwd)

env-up:
	@docker compose up -d sso-postgres

env-down:
	@docker compose down --remove-orphans sso-postgres port-forwarder
