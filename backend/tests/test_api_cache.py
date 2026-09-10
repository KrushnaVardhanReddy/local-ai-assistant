import sys
import os

sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
sys.modules['sounddevice'] = __import__('unittest.mock').mock.MagicMock()

import pytest
from fastapi.testclient import TestClient
<<<<<<< HEAD
from unittest.mock import patch, MagicMock
from app import app
import qa_cache
=======
from unittest.mock import patch, MagicMock, AsyncMock
from app import app
import qa_cache
import state
>>>>>>> origin/main

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
<<<<<<< HEAD
=======

def test_prewarm_cache():
    import json
    import app as app_module

    mock_llm_client = MagicMock()
    mock_response = '[{"question": "Q1", "answer": "A1"}, {"question": "Q2", "answer": "A2"}]'

    async def mock_stream(messages):
        yield mock_response

    mock_llm_client.stream = mock_stream
    app_module.llm_client = mock_llm_client
    app.llm_client = mock_llm_client

    with patch("app.qa_cache.store_bulk") as mock_store_bulk:
        mock_store_bulk.return_value = 2

        # Override the FastAPI app dependency or global
        # because TestClient starts lifespan which overrides app.llm_client
        with patch('app.LLMClient.from_config') as mock_from_config:
            async def return_mock(*args, **kwargs):
                return mock_llm_client
            mock_from_config.side_effect = return_mock

            with TestClient(app) as client:
                response = client.post("/api/cache/prewarm", json={"resume_text": "My Resume", "job_description": "My Job Description"})

                assert response.status_code == 200
                assert response.json() == {
                    "status": "success",
                    "message": "Cache warmed with 2 questions",
                    "stored_count": 2
                }
                mock_store_bulk.assert_called_once()
>>>>>>> origin/main
