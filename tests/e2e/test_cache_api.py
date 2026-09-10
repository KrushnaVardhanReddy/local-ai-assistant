import pytest
import httpx

BASE_URL = "http://127.0.0.1:8765"

class TestCacheAPI:

    def test_stats_returns_200(self):
        """GET /api/cache/stats should return 200 with correct shape."""
        resp = httpx.get(f"{BASE_URL}/api/cache/stats", timeout=5)
        assert resp.status_code == 200
        data = resp.json()
        assert "cached_pairs" in data
        assert "estimated_tokens_saved" in data
        assert isinstance(data["cached_pairs"], int)
        assert isinstance(data["estimated_tokens_saved"], int)

    def test_clear_cache_returns_200(self):
        """DELETE /api/cache should return 200 with deleted count."""
        resp = httpx.delete(f"{BASE_URL}/api/cache", timeout=5)
        assert resp.status_code == 200
        data = resp.json()
        assert "deleted" in data
        assert "message" in data
        assert isinstance(data["deleted"], int)

    def test_stats_after_clear_is_zero(self):
        """After DELETE /api/cache, GET /api/cache/stats should show 0 pairs."""
        httpx.delete(f"{BASE_URL}/api/cache", timeout=5)
        resp = httpx.get(f"{BASE_URL}/api/cache/stats", timeout=5)
        data = resp.json()
        assert data["cached_pairs"] == 0
