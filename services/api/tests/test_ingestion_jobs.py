from fastapi.testclient import TestClient

from app.deps.redis import get_redis
from app.main import app


class FakeRedis:
    def __init__(self) -> None:
        self.calls: list[tuple[str, str]] = []
        self.hashes: dict[str, dict[str, str]] = {}
        self.expirations: dict[str, int] = {}

    def lpush(self, queue: str, payload: str) -> int:
        self.calls.append((queue, payload))
        return 1

    def pipeline(self) -> "FakeRedis":
        return self

    def hset(self, key: str, mapping: dict[str, str]) -> int:
        current = self.hashes.setdefault(key, {})
        for k, v in mapping.items():
            current[str(k)] = str(v)
        return len(mapping)

    def hgetall(self, key: str) -> dict[str, str]:
        return dict(self.hashes.get(key, {}))

    def expire(self, key: str, ttl_seconds: int) -> bool:
        self.expirations[key] = int(ttl_seconds)
        return True

    def execute(self) -> list[object]:
        return []


def test_enqueue_job_generates_job_id_and_enqueues() -> None:
    fake = FakeRedis()
    app.dependency_overrides[get_redis] = lambda: fake

    try:
        client = TestClient(app)
        resp = client.post("/ingestion/jobs", json={"text": "hello", "source": "api"})
        assert resp.status_code == 200

        data = resp.json()
        assert data["queued"] is True
        assert data["queue"] == "ingestion:jobs"
        assert isinstance(data["job_id"], str)
        assert len(data["job_id"]) > 0

        assert len(fake.calls) == 1
        queue, payload = fake.calls[0]
        assert queue == "ingestion:jobs"
        assert '"job_id"' in payload
        assert '"payload"' in payload
        assert '"text": "hello"' in payload
    finally:
        app.dependency_overrides.clear()


def test_get_job_status_returns_404_until_processed() -> None:
    fake = FakeRedis()
    app.dependency_overrides[get_redis] = lambda: fake

    try:
        client = TestClient(app)
        resp = client.post("/ingestion/jobs", json={"text": "hello"})
        assert resp.status_code == 200
        job_id = resp.json()["job_id"]

        status_resp = client.get(f"/ingestion/jobs/{job_id}")
        assert status_resp.status_code == 404
    finally:
        app.dependency_overrides.clear()


def test_get_job_status_returns_processed_when_present() -> None:
    fake = FakeRedis()
    app.dependency_overrides[get_redis] = lambda: fake

    try:
        client = TestClient(app)
        job_id = "job-123"
        fake.hset(
            f"ingestion:job:{job_id}",
            mapping={
                "job_id": job_id,
                "status": "processed",
                "processed_at": "2026-05-28T00:00:00Z",
                "updated_at": "2026-05-28T00:00:00Z",
            },
        )

        status_resp = client.get(f"/ingestion/jobs/{job_id}")
        assert status_resp.status_code == 200
        data = status_resp.json()
        assert data["job_id"] == job_id
        assert data["status"] == "processed"
        assert data["processed_at"] == "2026-05-28T00:00:00Z"
        assert data["updated_at"] == "2026-05-28T00:00:00Z"
    finally:
        app.dependency_overrides.clear()


def test_enqueue_job_requires_non_empty_text() -> None:
    fake = FakeRedis()
    app.dependency_overrides[get_redis] = lambda: fake

    try:
        client = TestClient(app)
        resp = client.post("/ingestion/jobs", json={"text": "   "})
        assert resp.status_code == 422
        assert fake.calls == []
    finally:
        app.dependency_overrides.clear()
