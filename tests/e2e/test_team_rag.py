import pytest
import sys
import unittest.mock
from unittest.mock import patch, MagicMock

sys.path.append('backend')
from fastapi.testclient import TestClient
from backend.app import app
from backend.config import config

@pytest.fixture
def mock_supabase():
    # We will test without mocking supabase library directly in pytest,
    # and use the `TESTING_MOCK_SUPABASE` toggle in the endpoint.
    pass

@pytest.fixture
def mock_embedder():
    with patch('backend.rag.team_ingest.get_embedder') as mock_get_embedder:
        mock_emb = MagicMock()
        # Return a list of fake embeddings (one for each chunk)
        mock_emb.encode.return_value = MagicMock(tolist=lambda: [[0.1] * 384, [0.1] * 384])
        mock_get_embedder.return_value = mock_emb
        yield mock_emb

@pytest.fixture
def test_client():
    old_secret = config.SUPABASE_JWT_SECRET
    config.SUPABASE_JWT_SECRET = "test_secret"
    config.SUPABASE_URL = "http://mock.url"
    config.SUPABASE_SERVICE_KEY = "mock_key"

    # We must also inject test settings into team_ingest since it reads from config
    from backend.rag import team_ingest
    team_ingest.config.TESTING_MOCK_SUPABASE = True

    with TestClient(app) as client:
        yield client
    config.SUPABASE_JWT_SECRET = old_secret
    team_ingest.config.TESTING_MOCK_SUPABASE = False

@pytest.mark.asyncio
async def test_team_document_upload(test_client, mock_supabase, mock_embedder):
    from backend.rag import team_ingest

    with patch('backend.rag.team_ingest.get_user_org', new_callable=unittest.mock.AsyncMock) as mock_get_org:
        mock_get_org.return_value = "test-org"

        # We must also patch extract_text, chunk_text to avoid reading the dummy file
        with patch('backend.rag.team_ingest.extract_text', return_value="This is some text"):
            with patch('backend.rag.team_ingest.chunk_text', return_value=["This is some text"]):

                file_content = b'This is a test document. It has some text to chunk.'
                files = {'file': ('test.txt', file_content, 'text/plain')}

                # Send the x-test-user header which we added to bypass auth parsing in test mode
                team_ingest.config.mock_supabase_inserts = []
                response = test_client.post('/rag/team/ingest', files=files, headers={"x-test-user": "test-user", "x-test-org": "test-org"})

                assert response.status_code == 200, response.json()
                assert response.json()['ok'] is True

                # Check if our mock array got the inserts
                inserts = team_ingest.config.mock_supabase_inserts
                assert len(inserts) > 0

                found = False
                for insert_call in inserts:
                    if isinstance(insert_call, list) and len(insert_call) > 0:
                        if 'content' in insert_call[0] and insert_call[0].get('org_id') == 'test-org':
                            found = True
                            break
                    elif isinstance(insert_call, dict):
                        if insert_call.get('org_id') == 'test-org':
                            found = True
                            break

                assert found, "Did not find org_id in Supabase insert"

def test_team_retriever():
    from backend.rag.team_retriever import TeamRetriever

    # Mocking create_client in TeamRetriever module directly
    with patch('backend.rag.team_retriever.create_client') as mock_create:
        from backend.rag import team_retriever
        old_url = team_retriever.config.SUPABASE_URL
        old_key = team_retriever.config.SUPABASE_SERVICE_KEY
        team_retriever.config.SUPABASE_URL = "http://mock.url"
        team_retriever.config.SUPABASE_SERVICE_KEY = "mock_key"

        mock_supabase = MagicMock()
        mock_create.return_value = mock_supabase

        mock_rpc = MagicMock()
        mock_supabase.rpc.return_value = mock_rpc
        mock_rpc.execute.return_value = MagicMock(data=[
            {"content": "Match", "metadata": {"source_file": "test.txt", "chunk_index": 0}}
        ])

        retriever = TeamRetriever(org_id="test-org", query_embedding=[0.1]*384)

        assert retriever.supabase is not None

        results = retriever.retrieve(top_k=4)

        assert len(results) == 1
        assert results[0] == ("test.txt", "0", "Match")

        # Verify the RPC call included the org_id to prevent data leakage
        mock_supabase.rpc.assert_called_once()
        rpc_args = mock_supabase.rpc.call_args[0]
        assert rpc_args[0] == "match_team_documents"
        assert rpc_args[1]["p_org_id"] == "test-org"

        team_retriever.config.SUPABASE_URL = old_url
        team_retriever.config.SUPABASE_SERVICE_KEY = old_key
