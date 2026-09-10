import pytest
import io
import os
import sys
from fastapi.testclient import TestClient

# Add backend to path for imports
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "backend")))
from app import app
from config import config
from rag import team_ingest

@pytest.fixture
def client():
    with TestClient(app) as c:
        yield c

@pytest.fixture(autouse=True)
def mock_auth(monkeypatch):
    # Mock verify JWT and getting organization ID to simulate an enterprise user
    def mock_verify_supabase_jwt(*args, **kwargs):
        return {"sub": "test_user_id"}

    async def mock_get_user_org(*args, **kwargs):
        return "test_org_id"

    import auth
    monkeypatch.setattr("auth.get_user_org", mock_get_user_org)
    monkeypatch.setattr("auth.verify_supabase_jwt", mock_verify_supabase_jwt)

    # We must patch get_user_id directly in the imported module context!
    monkeypatch.setattr("rag.team_ingest.get_user_id", lambda t: "test_user_id")
    monkeypatch.setattr("auth.get_user_id", lambda t: "test_user_id")

    def mock_get_current_user_id(request=None):
        return "test_user_id"

    app.dependency_overrides[team_ingest.get_current_user_id] = mock_get_current_user_id

@pytest.fixture(autouse=True)
def mock_supabase_client(monkeypatch):
    class MockTable:
        def __init__(self, name):
            self.name = name

        def select(self, *args, **kwargs):
            return self

        def eq(self, *args, **kwargs):
            return self

        def delete(self, *args, **kwargs):
            return self

        def insert(self, *args, **kwargs):
            return self

        def execute(self):
            if self.name == "team_documents_meta":
                # Only return something when doing a get for testing team/list
                class MockResponse:
                    data = [{"filename": "team_test_doc.txt", "size_bytes": 100, "uploaded_at": "now", "uploaded_by": "test_user_id"}]
                return MockResponse()
            return None

    class MockClient:
        def table(self, table_name):
            return MockTable(table_name)

        def rpc(self, func_name, args):
            class MockQuery:
                def execute(self):
                    class MockResponse:
                        data = [{"content": "This is a team document about testing RAG.", "metadata": {"source_file": "team_test_doc.txt", "chunk_index": 0}}]
                    return MockResponse()
            return MockQuery()

    def mock_get_client():
        return MockClient()

    monkeypatch.setattr("rag.team_ingest.get_supabase_client", mock_get_client)
    monkeypatch.setattr("rag.retriever.TeamRetriever", lambda *args, **kwargs: MockTeamRetriever())

    class MockTeamRetriever:
        def retrieve(self, top_k=4):
             return [("team_test_doc.txt", "0", "This is a team document about testing RAG.")]


def test_team_rag_upload(client):
    file_content = b"This is a team document about testing RAG."
    files = {"file": ("team_test_doc.txt", file_content, "text/plain")}
    # Need to have some bearer token to pass initial header checks
    response = client.post("/rag/team/ingest", files=files, headers={"Authorization": "Bearer fake_token"})
    assert response.status_code == 200
    data = response.json()
    assert data["ok"] is True
    assert data["chunks"] > 0

def test_team_rag_list(client):
    response = client.get("/rag/team/list", headers={"Authorization": "Bearer fake_token"})
    assert response.status_code == 200
    data = response.json()
    assert "documents" in data
    assert len(data["documents"]) > 0
    assert data["documents"][0]["filename"] == "team_test_doc.txt"

def test_team_rag_delete(client):
    response = client.delete("/rag/team/document/team_test_doc.txt", headers={"Authorization": "Bearer fake_token"})
    assert response.status_code == 200
    data = response.json()
    assert data["ok"] is True

def test_team_rag_search_is_included(client, monkeypatch):
    # Test that standard search pulls from team_retriever if logged in
    async def mock_get_optional_user_id(*args, **kwargs):
        return "test_user_id"
    async def mock_get_user_org(*args, **kwargs):
        return "test_org_id"
    import app as app_mod
    monkeypatch.setattr("app.get_optional_user_id", mock_get_optional_user_id)

    import auth
    monkeypatch.setattr("auth.get_user_org", mock_get_user_org)

    # Override for this specific test
    app.dependency_overrides[app_mod.get_optional_user_id] = lambda: "test_user_id"

    # We must ensure config.SUPABASE_URL is set so TeamRetriever tries to connect
    monkeypatch.setattr(config, "SUPABASE_URL", "http://mock")
    monkeypatch.setattr(config, "SUPABASE_SERVICE_KEY", "mock")

    import rag.retriever as retriever
    context = retriever.retrieve("testing RAG", org_id="test_org_id")
    assert "team_test_doc.txt" in context

    # clean up
    del app.dependency_overrides[app_mod.get_optional_user_id]
