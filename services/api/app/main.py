from fastapi import FastAPI

from app.api.routes.ingestion import router as ingestion_router
from app.api.routes.meta import router as meta_router


def create_app() -> FastAPI:
    app = FastAPI(title="Document Intelligence Platform API")
    app.include_router(meta_router)
    app.include_router(ingestion_router)
    return app


app = create_app()
