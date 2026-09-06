import asyncio
import queue
import threading
import traceback
import sys
import re

import os
from fastapi import FastAPI, WebSocket, WebSocketDisconnect
from fastapi.staticfiles import StaticFiles
from fastapi.responses import JSONResponse
import uvicorn

import json
import httpx

from fastapi import Request, HTTPException, Depends, Response
from pydantic import BaseModel

from config import config, PROVIDER_CONFIG
from audio_listener import AudioListener, get_audio_devices
from auth import get_user_id, AuthError
from keys import key_store
from transcriber import Transcriber
from llm_client import LLMClient
from rag import ingestor
from rag import retriever
from rag.web_search import search_web
from smart_filter import SilenceBuffer, passes_filter

from contextlib import asynccontextmanager
from fastapi.middleware.cors import CORSMiddleware

@asynccontextmanager
async def lifespan(app: FastAPI):
    await startup_event()
    yield
    await shutdown_event()

app = FastAPI(lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(ingestor.router, prefix="/rag")
app.include_router(retriever.retriever_router, prefix="/rag")

frontend_dist = os.path.join(os.path.dirname(__file__), "../frontend/dist")
if os.path.exists(frontend_dist):
    app.mount("/helper", StaticFiles(directory=frontend_dist, html=True), name="helper")


# Global instances
transcriber = None
llm_client = None
listener = None
sync_queue = queue.Queue()
candidate_context: str = ""
preferred_language: str = ""

# PTT Mode global toggle
PTT_MODE = False

# Set of active per-connection asyncio.Queue objects for raw audio
active_ws_queues = set()

# Set of active per-connection asyncio.Queue objects for outbound JSON messages
active_outbound_queues = set()

def _audio_loop(listener: AudioListener, sync_queue: queue.Queue):
    """Background thread to capture audio chunks."""
    while True:
        try:
            chunk = listener.get_chunk()
            sync_queue.put(chunk)
        except Exception as e:
            print(f"Error in audio loop: {e}", file=sys.stderr)
            import time
            time.sleep(1)


async def _async_audio_bridge(sync_queue: queue.Queue, active_queues: set):
    """Bridges the sync thread to the async event loop without blocking."""
    while True:
        try:
            # Poll the sync queue without blocking
            chunk = sync_queue.get_nowait()
            for q in list(active_queues):
                q.put_nowait(chunk)
        except queue.Empty:
            pass
        except Exception as e:
            print(f"Error in async audio bridge: {e}", file=sys.stderr)

        # Yield control to the event loop
        await asyncio.sleep(0.01)


async def startup_event():
    global transcriber, llm_client, listener

    print("Starting up FastAPI application...", file=sys.stderr)

    # 1. Load Transcriber
    transcriber = Transcriber(
        model_size=config.STT_MODEL,
        device=config.STT_DEVICE,
        compute_type=config.STT_COMPUTE_TYPE,
        provider=config.STT_PROVIDER
    )
    # This blocks, but it's startup so it's fine. Could also use asyncio.to_thread.
    transcriber.load()

    # 2. Instantiate LLMClient
    llm_client = await LLMClient.from_config()
    print(f"LLM Client loaded with provider: {llm_client.provider}", file=sys.stderr)

    # 3. Start AudioListener
    listener = AudioListener(config.AUDIO_SAMPLE_RATE, config.AUDIO_CHUNK_SECONDS)
    listener.start()

    # 4. Start sync background thread
    threading.Thread(target=_audio_loop, args=(listener, sync_queue), daemon=True).start()

    # 5. Start async audio bridge task
    asyncio.create_task(_async_audio_bridge(sync_queue, active_ws_queues))


async def shutdown_event():
    if listener:
        listener.stop()


class LanguagePreference(BaseModel):
    language: str

@app.post("/api/language")
async def set_language(body: LanguagePreference):
    global preferred_language
    preferred_language = body.language
    return {"status": "ok", "language": preferred_language}

@app.get("/api/language")
async def get_language():
    return {"language": preferred_language}

@app.websocket("/ws")
async def ws_endpoint(websocket: WebSocket):
    await websocket.accept()

    connection_llm_client = llm_client

    if config.SUPABASE_JWT_SECRET:
        try:
            auth_msg_str = await asyncio.wait_for(websocket.receive_text(), timeout=5.0)
            auth_msg = json.loads(auth_msg_str)
            if auth_msg.get("type") != "auth" or "token" not in auth_msg or "machine_id" not in auth_msg:
                await websocket.close(code=1008, reason="Invalid auth message")
                return

            token = auth_msg["token"]
            machine_id = auth_msg["machine_id"]
            user_id = get_user_id(token)

            if not key_store:
                await websocket.close(code=1008, reason="Key store not configured")
                return

            try:
                await key_store.verify_user_profile(user_id, machine_id)
            except Exception as e:
                await websocket.close(code=1008, reason=str(e))
                return

            # Determine LLM provider from stored keys
            providers_to_check = ["openai", "groq", "gemini", "anthropic"]
            found_key = None
            found_provider = None

            if key_store:
                for prov in providers_to_check:
                    key = await key_store.get_key(user_id, prov)
                    if key:
                        found_key = key
                        found_provider = prov
                        break

            if found_key and found_provider:
                base_url = PROVIDER_CONFIG.get(found_provider, {}).get("url", "")
                connection_llm_client = LLMClient(
                    base_url=base_url,
                    model=config.LLM_MODEL,
                    api_key=found_key,
                    provider=found_provider
                )
            else:
                llm_config = config.resolved_llm()
                connection_llm_client = LLMClient(
                    base_url=llm_config.get("base_url", ""),
                    model=config.LLM_MODEL,
                    api_key=llm_config.get("api_key", ""),
                    provider=llm_config.get("provider", "")
                )

        except asyncio.TimeoutError:
            await websocket.close(code=1008, reason="Auth timeout")
            return
        except AuthError as e:
            await websocket.close(code=1008, reason=str(e))
            return
        except Exception as e:
            await websocket.close(code=1008, reason=f"Auth error: {str(e)}")
            return

    # Create a per-connection asyncio.Queue for audio chunks
    # Prevent HMR ghost clients from dual-processing audio by clearing old queues
    active_ws_queues.clear()
    ws_queue = asyncio.Queue()
    active_ws_queues.add(ws_queue)

    # Create a unified queue for outbound messages to this websocket
    outbound_queue = asyncio.Queue()
    active_outbound_queues.add(outbound_queue)

    is_streaming = asyncio.Event()  # Set when LLM is actively streaming
    pending_questions = asyncio.Queue()
    silence_buffer = SilenceBuffer()

    async def async_sender():
        try:
            while True:
                msg = await outbound_queue.get()
                if msg.get("type") == "token":
                    print(f"[DEBUG WS] Sending token to frontend: '{msg.get('text')}'", file=sys.stderr)
                await websocket.send_json(msg)
        except asyncio.CancelledError:
            pass
        except Exception as e:
            print(f"Error in async_sender: {e}", file=sys.stderr)

    async def async_receiver():
        try:
            while True:
                msg_str = await websocket.receive_text()
                try:
                    msg = json.loads(msg_str)
                    if msg.get("type") == "chat":
                        # We need to process the chat on the backend exactly once,
                        # but broadcast the user's transcript to all connected clients
                        # so they can see the message that was sent.
                        transcript_msg = {"type": "transcript", "text": msg.get("text", "")}

                        # Only put the LLM processing logic on ONE of the queues (the current one)
                        # so we don't trigger N concurrent LLM requests
                        ws_queue.put_nowait(msg)
                except Exception as e:
                    print(f"Error parsing incoming WS message: {e}", file=sys.stderr)
        except asyncio.CancelledError:
            pass
        except WebSocketDisconnect:
            pass
        except Exception as e:
            print(f"Error in async_receiver: {e}", file=sys.stderr)

    sender_task = asyncio.create_task(async_sender())
    receiver_task = asyncio.create_task(async_receiver())

    try:
        while True:
            try:
                # Await next audio chunk or chat message with 5s timeout
                chunk = await asyncio.wait_for(ws_queue.get(), timeout=5.0)

                if isinstance(chunk, dict) and chunk.get("type") == "chat":
                    transcript = chunk.get("text", "")
                else:
                    import numpy as np
                    import time
                    if not hasattr(websocket, "audio_buffer"):
                        websocket.audio_buffer = b""
                        websocket.last_active_time = time.monotonic()
                        websocket.last_live_transcript = ""
                        websocket.is_transcribing = False

                    audio_np = np.frombuffer(chunk, dtype=np.int16).astype(np.float32) / 32768.0
                    rms = np.sqrt(np.mean(audio_np**2))

                    if rms >= 0.01:
                        websocket.audio_buffer += chunk
                        websocket.last_active_time = time.monotonic()
                        
                        # Generate live transcript for the growing buffer (max 30s to prevent overflow)
                        if len(websocket.audio_buffer) > 16000 * 2 * 30: # 30 seconds max
                            websocket.audio_buffer = websocket.audio_buffer[-16000 * 2 * 30:]
                            
                        if len(websocket.audio_buffer) > 0 and not websocket.is_transcribing:
                            websocket.is_transcribing = True
                            try:
                                # Only transcribe the last 5 seconds for live preview, not the whole buffer
                                tail = websocket.audio_buffer[-16000 * 2 * 5:]
                                live_transcript = await asyncio.to_thread(transcriber.transcribe, tail)
                                if live_transcript and live_transcript != websocket.last_live_transcript:
                                    websocket.last_live_transcript = live_transcript
                                    for out_q in list(active_outbound_queues):
                                        out_q.put_nowait({"type": "transcript", "text": live_transcript})
                            finally:
                                websocket.is_transcribing = False
                        continue

                    # If silent and we have audio, check timeout
                    if websocket.audio_buffer and (time.monotonic() - websocket.last_active_time) >= config.SILENCE_THRESHOLD_SECONDS:
                        # Reuse the last live transcript if available, avoiding redundant Whisper call
                        if websocket.last_live_transcript:
                            assembled = websocket.last_live_transcript
                        else:
                            assembled = await asyncio.to_thread(transcriber.transcribe, websocket.audio_buffer)
                        websocket.audio_buffer = b""
                        websocket.last_active_time = time.monotonic()
                        websocket.last_live_transcript = ""
                        
                        print(f"[VAD] Assembled: '{assembled}'", file=sys.stderr)
                        
                        if not assembled:
                            continue
                            
                        should_send, reason = passes_filter(assembled)
                        if not should_send:
                            continue

                        transcript = assembled
                    else:
                        continue

                if not transcript:
                    continue

                # LAYER 2: Busy Guard
                if is_streaming.is_set():
                    print(f"[BUSY] Queued: '{transcript}'", file=sys.stderr)
                    pending_questions.put_nowait(transcript)
                    continue

                if transcript:
                    # Clear queue backlog
                    while not ws_queue.empty():
                        try:
                            ws_queue.get_nowait()
                        except:
                            pass

                    # Broadcast transcript to ALL connected clients and signal thinking state
                    for out_q in list(active_outbound_queues):
                        out_q.put_nowait({"type": "transcript", "text": transcript})
                        out_q.put_nowait({"type": "message_start"})

                    rag_context = ""
                    web_context = ""

                    loop = asyncio.get_event_loop()
                    tasks = []

                    if config.RAG_ENABLED:
                        tasks.append(loop.run_in_executor(None, retriever.retrieve, transcript))
                    else:
                        tasks.append(asyncio.sleep(0, result=""))

                    if config.WEB_SEARCH_ENABLED:
                        tasks.append(loop.run_in_executor(None, search_web, transcript))
                    else:
                        tasks.append(asyncio.sleep(0, result=""))

                    results = await asyncio.gather(*tasks)
                    rag_context = results[0]
                    web_context = results[1]

                    combined_context = "\n\n".join(filter(None, [rag_context, web_context]))

                    system_content = config.SYSTEM_PROMPT
                    if candidate_context:
                        system_content += f"\n\nCandidate profile: {candidate_context}"
                    if preferred_language:
                        system_content += f"\n\nIMPORTANT: Always provide all code examples in {preferred_language}. Do not use any other programming language for code unless the user explicitly asks."
                    if combined_context:
                        system_content += f"\n\n{combined_context}"
                        sources = list(dict.fromkeys(re.findall(r'\[Source: ([^\],]+)', combined_context)))
                        for out_q in list(active_outbound_queues):
                            out_q.put_nowait({"type": "rag_sources", "sources": sources})

                    messages = [
                        {"role": "system", "content": system_content},
                        {"role": "user", "content": transcript}
                    ]

                    is_streaming.set()
                    try:
                        print(f"[DEBUG] Calling LLM with {len(messages)} messages...", file=sys.stderr)
                        async for token in connection_llm_client.stream(messages):
                            for out_q in list(active_outbound_queues):
                                out_q.put_nowait({"type": "token", "text": token})
                        print(f"[DEBUG] Finished calling LLM.", file=sys.stderr)
                        for out_q in list(active_outbound_queues):
                            out_q.put_nowait({"type": "end"})
                    except Exception as e:
                        print(f"LLM Stream Error: {e}", file=sys.stderr)
                        for out_q in list(active_outbound_queues):
                            out_q.put_nowait({"type": "error", "message": f"LLM Error: {str(e)}"})
                    finally:
                        is_streaming.clear()
                        # Process any queued questions now that streaming is done
                        if not pending_questions.empty():
                            next_q = pending_questions.get_nowait()
                            # Put it back in the main queue to be processed as a "chat" bypass
                            ws_queue.put_nowait({"type": "chat", "text": next_q})

            except asyncio.TimeoutError:
                # Keep-alive ping
                # Preventing the WebSocket from closing on long silences
                outbound_queue.put_nowait({"type": "ping"})

    except WebSocketDisconnect:
        print("WebSocket disconnected.", file=sys.stderr)
    except Exception as e:
        print(f"WebSocket Error: {e}", file=sys.stderr)
        traceback.print_exc()
        try:
            await websocket.send_json({"type": "error", "message": str(e)})
        except:
            pass
    finally:
        sender_task.cancel()
        receiver_task.cancel()
        if ws_queue in active_ws_queues:
            active_ws_queues.remove(ws_queue)
        if outbound_queue in active_outbound_queues:
            active_outbound_queues.remove(outbound_queue)


class DeviceModel(BaseModel):
    device_id: int | None
    is_loopback: bool

@app.get("/api/audio/devices")
async def get_audio_devices_endpoint():
    return get_audio_devices()

@app.post("/api/audio/device")
async def set_audio_device(body: DeviceModel):
    global listener
    if listener:
        listener.stop()

    listener = AudioListener(
        config.AUDIO_SAMPLE_RATE,
        config.AUDIO_CHUNK_SECONDS,
        device=body.device_id,
        is_loopback=body.is_loopback
    )
    listener.start()
    return {"status": "success", "device_id": body.device_id, "is_loopback": body.is_loopback}


class PTTModeModel(BaseModel):
    enabled: bool

@app.post("/ptt/mode")
async def toggle_ptt_mode(body: PTTModeModel):
    global PTT_MODE
    PTT_MODE = body.enabled
    if PTT_MODE:
        if listener:
            listener.pause()
    else:
        if listener:
            listener.resume()
    return {"status": "success", "ptt_mode": PTT_MODE}

@app.post("/ptt/start")
async def ptt_start():
    if PTT_MODE and listener:
        listener.resume()
    return {"status": "started"}

@app.post("/ptt/stop")
async def ptt_stop():
    if PTT_MODE and listener:
        listener.pause()
    return {"status": "stopped"}

@app.post("/history/clear")
async def clear_history():
    # Because LLMClient builds its message history opaquely and ephemerally per request
    # (using transcript + system_prompt only) and holds no persistent list of conversational turns,
    # simply returning 'cleared' fulfills the panic clear structural requirement.
    return {"status": "cleared"}

class VisionModel(BaseModel):
    image_base64: str

@app.post("/vision/analyze")
async def analyze_vision(body: VisionModel):
    async def process_vision():
        vision_client = await LLMClient.from_config(is_vision=True)
        prompt = "Extract any coding problems, technical questions, or architecture diagrams from this screenshot. Provide a structured approach, pseudocode, and edge cases. Do not write the full code."

        # Broadcast message start to all connected clients
        for q in list(active_outbound_queues):
            q.put_nowait({"type": "message_start"})

        try:
            async for token in vision_client.chat_vision(body.image_base64, prompt):
                for q in list(active_outbound_queues):
                    q.put_nowait({"type": "token", "text": token})

            for q in list(active_outbound_queues):
                q.put_nowait({"type": "end"})
        except Exception as e:
            print(f"Vision Stream Error: {e}", file=sys.stderr)
            for q in list(active_outbound_queues):
                q.put_nowait({"type": "error", "message": f"Vision LLM Error: {str(e)}"})

    # Start the processing as a background task so the POST endpoint returns immediately
    asyncio.create_task(process_vision())
    return {"status": "processing started"}


class ResumeModel(BaseModel):
    text: str

@app.post("/api/resume/extract")
async def extract_resume(body: ResumeModel):
    global candidate_context, preferred_language
    if not llm_client:
        raise HTTPException(status_code=503, detail="LLM Client not initialized")

    system_prompt = (
        "You are a resume parser. Extract the candidate's key technical skills, "
        "years of experience, notable projects, and target role from the resume text. "
        "Be concise. Output a JSON object with exactly two keys:\n"
        "  - 'summary': A single paragraph of max 150 words describing the candidate.\n"
        "  - 'primary_language': The single most prominent programming language from\n"
        "    the resume (e.g. 'Python', 'JavaScript', 'Java', 'Go', 'C++', 'TypeScript').\n"
        "    If no clear language is found, return an empty string.\n"
        "Respond ONLY with valid JSON. No markdown, no extra text."
    )
    messages = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": body.text}
    ]

    result = ""
    try:
        async for token in llm_client.stream(messages):
            result += token
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"LLM Error: {str(e)}")

    llm_response_text = result.strip()

    # Check if the output is wrapped in a markdown code block and strip it
    if llm_response_text.startswith("```json"):
        llm_response_text = llm_response_text[7:]
    elif llm_response_text.startswith("```"):
        llm_response_text = llm_response_text[3:]
    if llm_response_text.endswith("```"):
        llm_response_text = llm_response_text[:-3]
    llm_response_text = llm_response_text.strip()

    try:
        result_json = json.loads(llm_response_text)
        candidate_context = result_json.get("summary", llm_response_text)
        detected_language = result_json.get("primary_language", "")
        if detected_language and not preferred_language:
            preferred_language = detected_language
    except (json.JSONDecodeError, KeyError):
        candidate_context = llm_response_text
        detected_language = ""

    return {
        "status": "success",
        "summary": candidate_context,
        "detected_language": detected_language
    }

@app.get("/api/resume/context")
async def get_resume_context():
    return {"context": candidate_context}


class PromptModel(BaseModel):
    prompt: str

@app.get("/api/system_prompt")
async def get_system_prompt():
    return {"prompt": config.SYSTEM_PROMPT}

@app.post("/api/system_prompt")
async def set_system_prompt(body: PromptModel):
    config.SYSTEM_PROMPT = body.prompt
    return {"status": "success"}

class KeyModel(BaseModel):
    api_key: str

async def get_current_user_id(request: Request) -> str:
    auth_header = request.headers.get("Authorization")
    if not auth_header or not auth_header.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Missing or invalid Authorization header")
    token = auth_header.split(" ")[1]
    try:
        return get_user_id(token)
    except AuthError as e:
        raise HTTPException(status_code=401, detail=str(e))

@app.post("/api/keys/{provider}")
async def save_api_key(provider: str, body: KeyModel, user_id: str = Depends(get_current_user_id)):
    if not key_store:
        raise HTTPException(status_code=500, detail="Key store not configured")
    await key_store.save_key(user_id, provider, body.api_key)
    return {"message": "Key saved successfully"}

@app.delete("/api/keys/{provider}")
async def delete_api_key(provider: str, user_id: str = Depends(get_current_user_id)):
    if not key_store:
        raise HTTPException(status_code=500, detail="Key store not configured")
    await key_store.delete_key(user_id, provider)
    return Response(status_code=204)

@app.post("/api/keys/{provider}/test")
async def test_api_key(provider: str, body: KeyModel, user_id: str = Depends(get_current_user_id)):
    base_url = PROVIDER_CONFIG.get(provider, {}).get("url", "")
    temp_client = LLMClient(base_url=base_url, model=config.LLM_MODEL, api_key=body.api_key, provider=provider)
    is_ok = await temp_client.health_check()
    return {"ok": is_ok}

@app.get("/health")
async def health_check():
    return {
        "status": "ok",
        "llm_provider": config.LLM_PROVIDER,
        "auth_enabled": bool(config.SUPABASE_JWT_SECRET)
    }

@app.get("/rag/status")
async def rag_status():
    return {"enabled": config.RAG_ENABLED}

@app.post("/rag/toggle")
async def rag_toggle():
    config.RAG_ENABLED = not config.RAG_ENABLED
    return {"enabled": config.RAG_ENABLED}

@app.get("/web_search/status")
async def web_search_status():
    return {"enabled": config.WEB_SEARCH_ENABLED}

@app.post("/web_search/toggle")
async def web_search_toggle():
    config.WEB_SEARCH_ENABLED = not config.WEB_SEARCH_ENABLED
    return {"enabled": config.WEB_SEARCH_ENABLED}


if __name__ == "__main__":
    uvicorn.run("app:app", host=config.WS_HOST, port=config.WS_PORT, reload=False)
