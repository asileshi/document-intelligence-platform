# API service (FastAPI)

Minimal FastAPI service.

## Local run (without Docker)

- `python -m venv .venv && source .venv/bin/activate`
- `pip install -r requirements.txt`
- `uvicorn app.main:app --reload --port 8000`

## Docker

This service is intended to run via the repo root `docker-compose.yml`.
