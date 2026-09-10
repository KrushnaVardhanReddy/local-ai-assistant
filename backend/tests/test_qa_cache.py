import sys
import os

# To avoid singleton duplication and module caching mismatches
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
os.environ["PYTHONPATH"] = "/app/backend" + (":" + os.environ.get("PYTHONPATH", "") if "PYTHONPATH" in os.environ else "")

import pytest
from unittest.mock import patch, MagicMock

import qa_cache

def test_cache_hit_and_miss():
    # Mock LocalIntelligence since it requires a real model file normally
    with patch('qa_cache.get_local_intelligence') as mock_get_li:
        mock_li = MagicMock()
        mock_get_li.return_value = mock_li

        # When embedding Q1
        def fake_encode(q):
            if "meaning of life" in q:
                return [1.0, 0.0, 0.0]
            else:
                return [0.0, 1.0, 0.0]

        mock_li.encode.side_effect = fake_encode

        qa_cache.clear()

        # Initial lookup should be a miss
        assert qa_cache.lookup("What is the meaning of life?") is None

        # Store a response
        qa_cache.store("What is the meaning of life?", "42")

        # Lookup should now be a hit
        assert qa_cache.lookup("What is the meaning of life?") == "42"

        # Different question should be a miss
        assert qa_cache.lookup("Different question") is None
