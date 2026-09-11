import pytest
from fastapi.testclient import TestClient

import sys
sys.path.append('backend')
from backend.app import app

class TestCacheAPI:

    @pytest.fixture(autouse=True)
    def setup_client(self):
        self.client = TestClient(app)

    def test_stats_returns_200(self):
        """GET /api/cache/stats should return 200 with correct shape."""
        resp = self.client.get("/api/cache/stats")
        assert resp.status_code == 200
        data = resp.json()
        assert "cached_pairs" in data
        assert "estimated_tokens_saved" in data
        assert isinstance(data["cached_pairs"], int)
        assert isinstance(data["estimated_tokens_saved"], int)

    def test_clear_cache_returns_200(self):
        """DELETE /api/cache should return 200 with deleted count."""
        resp = self.client.delete("/api/cache")
        assert resp.status_code == 200
        data = resp.json()
        assert "deleted" in data
        assert "message" in data
        assert isinstance(data["deleted"], int)

    def test_stats_after_clear_is_zero(self):
        """After DELETE /api/cache, GET /api/cache/stats should show 0 pairs."""
        self.client.delete("/api/cache")
        resp = self.client.get("/api/cache/stats")
        data = resp.json()
        assert data["cached_pairs"] == 0
