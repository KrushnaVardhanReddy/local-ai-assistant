import pytest
from fastapi.testclient import TestClient
from backend.app import app

@pytest.fixture
def test_client():
    with TestClient(app) as client:
        yield client

def test_tailor_resume(test_client):
    from unittest.mock import patch, AsyncMock

    # We use app_module to patch the global llm_client
    import backend.app as app_module

    with patch.object(app_module, "llm_client", create=True) as mock_llm_client:
        mock_llm_client.return_value = True

        # We need to mock resume_builder.generate_tailored_resume inside app_module
        # since app.py imports resume_builder
        with patch.object(app_module.resume_builder, "generate_tailored_resume", new_callable=AsyncMock) as mock_generate:
            mock_generate.return_value = "Tailored Resume"

            response = test_client.post("/resume/tailor", json={"base_resume": "I worked as SWE at Google.", "job_description": "Looking for SWE at Amazon."})
            assert response.status_code == 200
            assert response.json()["markdown"] == "Tailored Resume"
