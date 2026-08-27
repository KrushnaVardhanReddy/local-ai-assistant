import pytest
import websockets
import json
import asyncio
from fastapi.testclient import TestClient
from backend.app import app

@pytest.mark.asyncio
async def test_ws_health_frame():
    # Use fastapi's TestClient which natively supports websockets
    with TestClient(app) as client:
        with client.websocket_connect("/ws") as websocket:
            # We don't send anything. We expect a ping within 5 seconds + some buffer.
            # However, TestClient.websocket_connect blocks on receive(), so we use receive_json
            data = websocket.receive_json()
            assert data.get("type") == "ping"
