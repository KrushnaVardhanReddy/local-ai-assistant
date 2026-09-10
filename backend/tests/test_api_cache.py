import sys
import os

sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
sys.modules['sounddevice'] = __import__('unittest.mock').mock.MagicMock()

import pytest
from fastapi.testclient import TestClient
from unittest.mock import patch, MagicMock
from app import app
import qa_cache

def test_get_cache_stats():
    with patch("app.qa_cache.stats") as mock_stats:
        mock_stats.return_value = {"cached_pairs": 10, "estimated_tokens_saved": 2560}

        with TestClient(app) as client:
            response = client.get("/api/cache/stats")

            assert response.status_code == 200
            assert response.json() == {"cached_pairs": 10, "estimated_tokens_saved": 2560}
            mock_stats.assert_called_once()

def test_delete_cache():
    with patch("app.qa_cache.clear") as mock_clear:
        mock_clear.return_value = 5

        with TestClient(app) as client:
            response = client.delete("/api/cache")

            assert response.status_code == 200
            assert response.json() == {"deleted": 5, "message": "Cleared 5 cached Q&A pairs."}
            mock_clear.assert_called_once()
