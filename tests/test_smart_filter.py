import pytest
import time
import os

# We need to import config first before anything in smart_filter is loaded
# so that we can patch its values for testing if needed.
from config import config
from smart_filter import passes_filter, SilenceBuffer, is_question, is_filler

def test_is_filler():
    assert is_filler("okay")
    assert is_filler("OKAY.")
    assert is_filler("mm-hmm")
    assert is_filler("thanks.")
    assert not is_filler("thanks for the help")
    assert not is_filler("what is react?")

def test_is_question():
    assert is_question("What is react?")
    assert is_question("Can you explain this?")
    assert is_question("difference between var and let")
    assert is_question("How does this work")
    assert is_question("how can i fix this")
    assert not is_question("I am just talking.")
    assert not is_question("The sun is bright.")

def test_passes_filter_too_short():
    config.MIN_WORDS = 4
    config.SMART_FILTER_ENABLED = True
    should_send, reason = passes_filter("Too short phrase")
    assert not should_send
    assert reason == "too_short"

def test_passes_filter_filler():
    # Set it lower than the actual word count of the filler to hit the filler logic
    config.MIN_WORDS = 1
    config.SMART_FILTER_ENABLED = True
    should_send, reason = passes_filter("Okay.")
    assert not should_send
    assert reason == "filler"

def test_passes_filter_not_question():
    config.MIN_WORDS = 4
    config.SMART_FILTER_ENABLED = True
    should_send, reason = passes_filter("This is just a long sentence that is not a question.")
    assert not should_send
    assert reason == "not_question"

def test_passes_filter_accepted():
    config.MIN_WORDS = 4
    config.SMART_FILTER_ENABLED = True
    should_send, reason = passes_filter("What is the difference between React and Vue?")
    assert should_send
    assert reason == "accepted"

def test_passes_filter_disabled():
    config.SMART_FILTER_ENABLED = False
    # Even if it's too short and not a question, it should pass if filter is disabled
    should_send, reason = passes_filter("yes")
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
