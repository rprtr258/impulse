SHELL := /usr/bin/env bash

.DEFAULT_GOAL := dev

.PHONY: db_init
db_init:
	@echo "Initializing database..."

.PHONY: db_drop
db_drop:
	@echo "Dropping database..."
	rm -f dist/*

.PHONY: db_reinit
db_reinit:
	@echo "Reinitializing database..."
	${MAKE} db_drop
	${MAKE} db_init

.PHONY: run
run:
	go mod tidy
	go run cmd/main.go

.PHONY: dev
dev:
	reflex -sr '(\.go$$)' -- $(MAKE) run
