import pytest
import time
import os

# We need to import config first before anything in smart_filter is loaded
# so that we can patch its values for testing if needed.
from config import config
from smart_filter import passes_filter, SilenceBuffer, is_filler

def test_is_filler():
    assert is_filler("okay")
    assert is_filler("OKAY.")
    assert is_filler("mm-hmm")
    assert is_filler("thanks.")
    assert not is_filler("thanks for the help")
    assert not is_filler("what is react?")

@pytest.mark.asyncio
async def test_passes_filter_too_short():
    config.MIN_WORDS = 4
    config.SMART_FILTER_ENABLED = True
    should_send, reason = await passes_filter("Too short phrase")
    assert not should_send
    assert reason == "too_short"

@pytest.mark.asyncio
async def test_passes_filter_filler():
    # Set it lower than the actual word count of the filler to hit the filler logic
    config.MIN_WORDS = 1
    config.SMART_FILTER_ENABLED = True
    should_send, reason = await passes_filter("Okay.")
    assert not should_send
    assert reason == "filler"

@pytest.mark.asyncio
async def test_passes_filter_not_question():
    from unittest.mock import patch
    import asyncio
    config.MIN_WORDS = 4
    config.SMART_FILTER_ENABLED = True
    with patch("smart_filter.get_local_intelligence") as mock_get_li:
        mock_li = mock_get_li.return_value

        async def mock_is_question(text):
            return False

        mock_li.is_question = mock_is_question
        should_send, reason = await passes_filter("This is just a long sentence that is not a question.")
        assert not should_send
        assert reason == "not_question"

@pytest.mark.asyncio
async def test_passes_filter_accepted():
    config.MIN_WORDS = 4
    config.SMART_FILTER_ENABLED = True
    should_send, reason = await passes_filter("What is the difference between React and Vue?")
    assert should_send
    assert reason == "accepted"

@pytest.mark.asyncio
async def test_passes_filter_disabled():
    config.SMART_FILTER_ENABLED = False
    # Even if it's too short and not a question, it should pass if filter is disabled
    should_send, reason = await passes_filter("yes")
    assert should_send
    assert reason == "filter_disabled"

def test_silence_buffer():
    config.SILENCE_THRESHOLD_SECONDS = 0.1
    buffer = SilenceBuffer()

    assert buffer.is_empty()
    assert not buffer.is_silent() # empty buffer is not silent

    buffer.add("Hello")
    assert not buffer.is_empty()
    assert not buffer.is_silent() # just added, not silent

    buffer.add("World")
    assert not buffer.is_silent() # just added, not silent

    time.sleep(0.15)
    assert buffer.is_silent()

    flushed = buffer.flush()
    assert flushed == "Hello World"
    assert buffer.is_empty()
    assert not buffer.is_silent() # empty again
