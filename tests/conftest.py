import pytest
import asyncio
import httpx
from backend.app import app

@pytest.fixture
async def app_client():
    async with httpx.AsyncClient(transport=httpx.ASGITransport(app=app), base_url="http://testserver") as client:
        yield client
