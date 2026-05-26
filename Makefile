.DEFAULT_GOAL := help

COMPOSE := docker compose -f docker-compose.yml

.PHONY: help
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Docker Compose:"
	@echo "  up            Build & start services"
	@echo "  down          Stop services and remove volumes"
	@echo "  ps            Show running services"
	@echo "  logs          Tail service logs"
	@echo ""
	@echo "API (FastAPI):"
	@echo "  api-install   Install API runtime + dev dependencies"
	@echo "  api-lint      Run ruff"
	@echo "  api-format    Format with black (modifies files)"
	@echo "  api-fmt-check Check formatting with black"
	@echo "  api-test      Run pytest"
	@echo "  api-check     Run lint + format check + tests"
	@echo ""
	@echo "CI-like:"
	@echo "  ci            Run api-check + compose smoke checks"

.PHONY: up down ps logs build restart clean

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down -v

ps:
	$(COMPOSE) ps

logs:
	$(COMPOSE) logs -f --tail=100

build:
	$(COMPOSE) build

restart:
	$(COMPOSE) restart

clean:
	$(COMPOSE) down -v

.PHONY: api-install api-lint api-format api-fmt-check api-test api-check

api-install:
	python3 -m pip install --upgrade pip
	python3 -m pip install -r services/api/requirements.txt -r services/api/requirements-dev.txt

api-lint:
	ruff check services/api

api-format:
	black services/api

api-fmt-check:
	black --check services/api

api-test:
	cd services/api && pytest -q

api-check: api-lint api-fmt-check api-test

.PHONY: wait-qdrant wait-api ci

wait-qdrant:
	@i=1; \
	while [ $$i -le 30 ]; do \
		if curl -fsS http://localhost:6333/healthz >/dev/null; then \
			echo "Qdrant is ready"; \
			exit 0; \
		fi; \
		echo "Waiting for Qdrant... ($$i/30)"; \
		i=$$((i+1)); \
		sleep 2; \
	done; \
	echo "Qdrant did not become ready in time"; \
	$(COMPOSE) ps; \
	$(COMPOSE) logs qdrant; \
	exit 1

wait-api:
	@i=1; \
	while [ $$i -le 30 ]; do \
		if curl -fsS http://localhost:8000/health >/dev/null; then \
			echo "API is ready"; \
			exit 0; \
		fi; \
		echo "Waiting for API... ($$i/30)"; \
		i=$$((i+1)); \
		sleep 2; \
	done; \
	echo "API did not become ready in time"; \
	$(COMPOSE) ps; \
	$(COMPOSE) logs api; \
	exit 1

ci: api-install api-check
	$(COMPOSE) up -d --build
	$(MAKE) wait-qdrant
	$(MAKE) wait-api
	$(COMPOSE) down -v
