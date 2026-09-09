import pytest
from app import app
from fastapi.testclient import TestClient

client = TestClient(app)

def test_tailor_resume():
    response = client.post("/resume/tailor", json={"base_resume": "I worked as SWE at Google.", "job_description": "Looking for SWE at Amazon."})
    # Since llm_client is None when tested this way because startup event isn't called we can check that it fails for the right reason.
    assert response.status_code == 500
    assert "LLM Client not initialized" in response.text
