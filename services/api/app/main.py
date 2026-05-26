from fastapi import FastAPI

app = FastAPI(title="Document Intelligence Platform API")


@app.get("/health")
def health() -> dict:
    return {"status": "ok"}


@app.get("/version")
def version() -> dict:
    return {"service": "api", "version": "0.0.0"}
