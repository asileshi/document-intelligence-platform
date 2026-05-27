from __future__ import annotations

from functools import lru_cache

import redis

from app.core.config import Settings, get_settings


@lru_cache(maxsize=1)
def _client(settings: Settings) -> redis.Redis:
    return redis.Redis.from_url(settings.redis_url, decode_responses=True)


def get_redis() -> redis.Redis:
    settings = get_settings()
    return _client(settings)
