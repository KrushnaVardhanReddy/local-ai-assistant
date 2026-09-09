
I will refactor `backend/app.py` into modular FastAPI routers and a shared state module (`backend/state.py`) according to the task description. Here is the step-by-step plan:

1. **Create `backend/state.py`**:
   - Extract global variables from `app.py`: `transcriber`, `llm_client`, `listener`, `sync_queue`, `candidate_context`, `preferred_language`, `PTT_MODE`, `active_ws_queues`, `active_outbound_queues`.
   - Add appropriate type hints and default values to `backend/state.py`.
   - Important: When these variables are mutated in other files, they should be imported as `import state` and mutated as `state.PTT_MODE = ...`, rather than `from state import PTT_MODE`.

2. **Create `backend/routes/settings.py`**:
   - Move endpoints: `/api/language`, `/api/interview_language`, `/api/system_prompt`, `/config/job-description`, `/api/audio/devices`, `/api/audio/device`.
   - Also move endpoints: `/api/keys/{provider}`, `/api/keys/{provider}/test`, `/health`, `/rag/status`, `/rag/toggle`, `/web_search/status`, `/web_search/toggle`, `/api/internal/billing/update_plan`. Let's create `routes/keys.py` and `routes/system.py` or just group them sensibly. The task asks for settings specifically, but didn't list keys/health/etc in the groups, I'll put keys and other settings in `settings.py` or separate files if appropriate, but wait, the instructions specify:
     - **`routes/settings.py`**: `/api/language`, `/api/interview_language`, `/api/system_prompt`, `/config/job-description`, `/api/audio/devices`, `/api/audio/device`. Let's strictly follow this. I'll add the remaining ones (`/api/keys/*`, `/health`, `/rag/*`, `/web_search/*`) to `routes/settings.py` or `routes/system.py`. Let's put keys and configs in settings or keep them in `app.py`? Actually I'll create `routes/keys.py` and `routes/system.py` for remaining stuff to be clean, or put them in `settings.py`. Let's put them in `settings.py`.

3. **Create `backend/routes/audio.py`**:
   - Move endpoints: `/ptt/mode`, `/ptt/start`, `/ptt/stop`.

4. **Create `backend/routes/vision.py`**:
   - Move endpoint: `/vision/analyze`.

5. **Create `backend/routes/resume.py`**:
   - Move endpoints: `/resume/tailor`, `/api/resume/extract`, `/api/resume/context`.

6. **Create `backend/routes/history.py`**:
   - Move endpoints: `/history/clear`, `/session/end`, `/session/clear`, `/session/history`, `/session/email-draft`. The instructions say `/history/clear`, I will also move the other `/session/*` endpoints here or to a `routes/session.py`. I'll put them in `routes/history.py` as it makes sense.

7. **Create `backend/routes/ws.py`**:
   - Move the `@app.websocket("/ws")` endpoint.
   - Move all helper async tasks directly used by `ws_endpoint`: `async_receiver`, `async_sender`, `session_validator`, `llm_processor`, `audio_chunk_processor`, `video_frame_processor`, `mock_interviewer_task`, `is_reasoning_model`, etc.

8. **Refactor `backend/app.py`**:
   - Import `backend.state as state`.
   - Retain `lifespan` function which starts the threads.
   - Retain `CORSMiddleware` and `app.mount`.
   - Update `_audio_loop` and `_async_audio_bridge` to use `state.sync_queue` and `state.active_ws_queues`.
   - Remove extracted routes.
   - Add `app.include_router(...)` for all new routers.

9. **Verification**:
   - Run `python3 -m py_compile backend/app.py backend/routes/*.py` and `backend/state.py` to ensure no syntax errors.
   - Complete pre-commit instructions.
