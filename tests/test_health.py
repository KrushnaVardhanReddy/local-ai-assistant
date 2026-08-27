import pytest

@pytest.mark.asyncio
async def test_health_ok(app_client):
    response = await app_client.get("/health")
    assert response.status_code == 200
    assert response.json()["status"] == "ok"
