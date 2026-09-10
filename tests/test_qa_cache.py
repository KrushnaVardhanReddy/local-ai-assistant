import sys
import os
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '../backend')))

import pytest
import tempfile
from unittest.mock import patch, MagicMock

@pytest.fixture
def temp_chroma(tmp_path):
    """Provides a temporary ChromaDB directory for isolated tests."""
    return str(tmp_path / "chroma_test")

class TestQACache:

    def test_lookup_returns_none_when_smollm2_disabled(self, temp_chroma):
        """Cache lookup returns None when SmolLM2 is disabled (no embeddings)."""
        mock_li = MagicMock()
        mock_li.encode.return_value = []  # disabled state
        with patch("qa_cache.get_local_intelligence", return_value=mock_li), \
             patch("qa_cache.config") as mock_cfg:
            mock_cfg.CHROMA_DIR = temp_chroma
            import qa_cache
            result = qa_cache.lookup("What is a closure?")
            assert result is None

    def test_store_and_lookup_returns_cached_answer(self, temp_chroma):
        """Storing a Q&A pair and then looking up a similar question returns the answer."""
        fake_vector = [0.1] * 384
        mock_li = MagicMock()
        mock_li.encode.return_value = fake_vector
        import qa_cache
        qa_cache._collection = None
        qa_cache._client = None
        with patch("qa_cache.get_local_intelligence", return_value=mock_li), \
             patch("qa_cache.config") as mock_cfg:
            mock_cfg.CHROMA_DIR = temp_chroma
            qa_cache.store("What is a closure?", "A closure is a function with access to its outer scope.")
            result = qa_cache.lookup("What is a closure?")
            assert result == "A closure is a function with access to its outer scope."

    def test_clear_empties_collection(self, temp_chroma):
        """clear() removes all entries and stats() returns 0 after clearing."""
        fake_vector = [0.2] * 384
        mock_li = MagicMock()
        mock_li.encode.return_value = fake_vector
        import qa_cache
        qa_cache._collection = None
        qa_cache._client = None
        with patch("qa_cache.get_local_intelligence", return_value=mock_li), \
             patch("qa_cache.config") as mock_cfg:
            mock_cfg.CHROMA_DIR = temp_chroma
            qa_cache.store("Q1", "A1")
            qa_cache.store("Q2", "A2")
            count = qa_cache.clear()
            assert count == 2
            assert qa_cache.stats()["cached_pairs"] == 0

    def test_stats_reports_correct_count(self, temp_chroma):
        """stats() returns accurate cached_pairs count."""
        fake_vector = [0.3] * 384
        mock_li = MagicMock()
        mock_li.encode.return_value = fake_vector
        import qa_cache
        qa_cache._collection = None
        qa_cache._client = None
        with patch("qa_cache.get_local_intelligence", return_value=mock_li), \
             patch("qa_cache.config") as mock_cfg:
            mock_cfg.CHROMA_DIR = temp_chroma
            qa_cache.store("Question 1", "Answer 1")
            stats = qa_cache.stats()
            assert stats["cached_pairs"] == 1
            assert stats["estimated_tokens_saved"] > 0
