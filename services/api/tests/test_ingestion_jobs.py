from fastapi.testclient import TestClient

from app.deps.redis import get_redis
from app.main import app


class FakeRedis:
    def __init__(self) -> None:
        self.calls: list[tuple[str, str]] = []

    def lpush(self, queue: str, payload: str) -> int:
        self.calls.append((queue, payload))
        return 1


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
