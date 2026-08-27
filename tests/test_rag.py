import pytest
import io

@pytest.mark.asyncio
async def test_upload_txt(app_client):
    file_content = b"Hello world RAG test content"
    files = {"file": ("test.txt", file_content, "text/plain")}
    response = await app_client.post("/rag/upload", files=files)
    assert response.status_code == 200
    data = response.json()
    assert data["ok"] is True
    assert data["chunks"] > 0

@pytest.mark.asyncio
async def test_list_documents(app_client):
    response = await app_client.get("/rag/documents")
    assert response.status_code == 200
    data = response.json()
    assert "test.txt" in data.get("documents", [])

@pytest.mark.asyncio
async def test_search(app_client):
    response = await app_client.get("/rag/search?q=RAG")
    assert response.status_code == 200
    data = response.json()
    assert data["context"] != ""
    assert data["empty"] is False

@pytest.mark.asyncio
async def test_delete_document(app_client):
    response = await app_client.delete("/rag/document/test.txt")
    assert response.status_code == 200
    data = response.json()
    assert data["ok"] is True

@pytest.mark.asyncio
async def test_search_empty_after_delete(app_client):
    response = await app_client.get("/rag/search?q=RAG")
    assert response.status_code == 200
    data = response.json()
    assert data["empty"] is True
