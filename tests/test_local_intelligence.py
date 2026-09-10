import sys
import os
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '../backend')))

import pytest
from unittest.mock import patch, MagicMock

class TestLocalIntelligence:

    def test_is_complete_returns_true_when_disabled(self):
        """When SMOLLM2_ENABLED=false, is_complete always returns True."""
        with patch("local_intelligence.config") as mock_cfg:
            mock_cfg.SMOLLM2_ENABLED = False
            from local_intelligence import LocalIntelligence
            li = LocalIntelligence()
            assert li.is_complete("Can you tell me") is True

    def test_is_complete_returns_true_for_complete_sentence(self):
        """SmolLM2 returning COMPLETE should pass the text through."""
        mock_llm = MagicMock()
        mock_llm.return_value = {"choices": [{"text": "COMPLETE"}]}
        with patch("local_intelligence.config") as mock_cfg:
            try:
                import llama_cpp
                has_llama = True
            except ImportError:
                has_llama = False

            if has_llama:
                with patch("local_intelligence.Llama", return_value=mock_llm):
                    mock_cfg.SMOLLM2_ENABLED = True
                    mock_cfg.SMOLLM2_MODEL_PATH = "/fake/model.gguf"
                    from local_intelligence import LocalIntelligence
                    li = LocalIntelligence.__new__(LocalIntelligence)
                    li._enabled = True
                    li._llm = mock_llm
                    assert li.is_complete("What is a closure in JavaScript?") is True
            else:
                mock_cfg.SMOLLM2_ENABLED = True
                mock_cfg.SMOLLM2_MODEL_PATH = "/fake/model.gguf"
                from local_intelligence import LocalIntelligence
                li = LocalIntelligence.__new__(LocalIntelligence)
                li._enabled = True
                li._llm = mock_llm
                assert li.is_complete("What is a closure in JavaScript?") is True

    def test_is_complete_returns_false_for_incomplete_sentence(self):
        """SmolLM2 returning INCOMPLETE should block the text."""
        mock_llm = MagicMock()
        mock_llm.return_value = {"choices": [{"text": "INCOMPLETE"}]}
        from local_intelligence import LocalIntelligence
        li = LocalIntelligence.__new__(LocalIntelligence)
        li._enabled = True
        li._llm = mock_llm
        assert li.is_complete("Can you explain how to...") is False

    def test_encode_returns_empty_when_disabled(self):
        """When SMOLLM2_ENABLED=false, encode returns empty list."""
        from local_intelligence import LocalIntelligence
        li = LocalIntelligence.__new__(LocalIntelligence)
        li._enabled = False
        li._llm = None
        assert li.encode("some text") == []

    def test_is_complete_falls_back_on_exception(self):
        """If SmolLM2 raises an exception, is_complete returns True (safe fallback)."""
        mock_llm = MagicMock()
        mock_llm.side_effect = RuntimeError("model crashed")
        from local_intelligence import LocalIntelligence
        li = LocalIntelligence.__new__(LocalIntelligence)
        li._enabled = True
        li._llm = mock_llm
        assert li.is_complete("test") is True
