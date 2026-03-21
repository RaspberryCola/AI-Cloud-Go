SHELL := /bin/bash

.PHONY: build up down logs config

build:
	docker compose -f docker-compose.full.yml build

up:
	docker compose -f docker-compose.full.yml up -d --build

down:
	docker compose -f docker-compose.full.yml down

logs:
	docker compose -f docker-compose.full.yml logs -f --tail=200

config:
	docker compose -f docker-compose.full.yml config
