from __future__ import annotations

import json
from uuid import uuid4

from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel
import redis

from app.core.config import Settings, get_settings
from app.deps.redis import get_redis

router = APIRouter(prefix="/ingestion", tags=["ingestion"])


class EnqueueJobRequest(BaseModel):
    text: str
    source: str = "api"


class EnqueueJobResponse(BaseModel):
    job_id: str
    queued: bool
    queue: str


class JobStatusResponse(BaseModel):
    job_id: str
    status: str
    updated_at: str | None = None
    processed_at: str | None = None


@router.post("/jobs", response_model=EnqueueJobResponse)
def enqueue_job(
    body: EnqueueJobRequest,
    rdb: redis.Redis = Depends(get_redis),
    settings: Settings = Depends(get_settings),
) -> EnqueueJobResponse:
    if not body.text.strip():
        raise HTTPException(status_code=422, detail="text must not be empty")

    job_id = str(uuid4())
    payload = {
        "job_id": job_id,
        "source": body.source,
        "payload": {"text": body.text},
    }

    try:
        rdb.lpush(settings.redis_queue, json.dumps(payload))
    except redis.RedisError as exc:
        raise HTTPException(status_code=503, detail="redis unavailable") from exc

    return EnqueueJobResponse(job_id=job_id, queued=True, queue=settings.redis_queue)


@router.get("/jobs/{job_id}", response_model=JobStatusResponse)
def get_job_status(
    job_id: str,
    rdb: redis.Redis = Depends(get_redis),
    settings: Settings = Depends(get_settings),
) -> JobStatusResponse:
    status_key = f"{settings.redis_job_status_prefix}:{job_id}"
    try:
        data = rdb.hgetall(status_key)
    except redis.RedisError as exc:
        raise HTTPException(status_code=503, detail="redis unavailable") from exc

    if not data:
        raise HTTPException(status_code=404, detail="job not found")

    return JobStatusResponse(
        job_id=data.get("job_id", job_id),
        status=data.get("status", "unknown"),
        updated_at=data.get("updated_at"),
        processed_at=data.get("processed_at"),
    )
