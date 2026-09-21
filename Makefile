include .env

export

.PHONY: run

run:
	@go run ./cmd/sso --config config/local.yaml
