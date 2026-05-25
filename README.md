# Document Intelligence Platform

Monorepo for an AI document intelligence platform (RAG + ingestion workers).

## Local development (infrastructure)

1) Create local env file:

- `cp .env.example .env`

2) Start dependencies:

- `docker compose up -d`

3) Check health:

- Qdrant: http://localhost:6333/healthz
- Postgres: `docker compose exec postgres pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"`
- Redis: `docker compose exec redis redis-cli ping`

## Repo layout

- `frontend/` — Next.js UI
- `services/api/` — FastAPI service
- `services/ingestion-worker/` — Go worker service
- `shared/` — shared schemas/prompts/config
- `infrastructure/` — Docker/Nginx/scripts/terraform (optional later)
