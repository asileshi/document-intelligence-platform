from __future__ import annotations

from dataclasses import dataclass
import os


@dataclass(frozen=True)
class Settings:
    redis_url: str
    redis_queue: str
    redis_job_status_prefix: str


def get_settings() -> Settings:
    return Settings(
        redis_url=os.getenv("REDIS_URL", "redis://redis:6379/0"),
        redis_queue=os.getenv("REDIS_QUEUE", "ingestion:jobs"),
        redis_job_status_prefix=os.getenv("REDIS_JOB_STATUS_PREFIX", "ingestion:job"),
    )
