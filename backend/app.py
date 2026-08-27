import asyncio
import queue
import threading
import traceback
import sys
import re

from fastapi import FastAPI, WebSocket, WebSocketDisconnect
from fastapi.responses import JSONResponse
import uvicorn

import json
import httpx

from fastapi import Request, HTTPException, Depends, Response
from pydantic import BaseModel

from config import config, PROVIDER_CONFIG
from audio_listener import AudioListener
from auth import get_user_id, AuthError
from keys import key_store
from transcriber import Transcriber
from llm_client import LLMClient
from rag import ingestor
from rag import retriever

from contextlib import asynccontextmanager

@asynccontextmanager
async def lifespan(app: FastAPI):
    await startup_event()
    yield
    await shutdown_event()

app = FastAPI(lifespan=lifespan)
app.include_router(ingestor.router, prefix="/rag")
app.include_router(retriever.retriever_router, prefix="/rag")


# Global instances
transcriber = None
llm_client = None
listener = None
sync_queue = queue.Queue()

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
        compute_type=config.STT_COMPUTE_TYPE
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
    ws_queue = asyncio.Queue()
    active_ws_queues.add(ws_queue)

    # Create a unified queue for outbound messages to this websocket
    outbound_queue = asyncio.Queue()
    active_outbound_queues.add(outbound_queue)

    async def async_sender():
        try:
            while True:
                msg = await outbound_queue.get()
                await websocket.send_json(msg)
        except asyncio.CancelledError:
            pass
        except Exception as e:
            print(f"Error in async_sender: {e}", file=sys.stderr)

    sender_task = asyncio.create_task(async_sender())

    try:
        while True:
            try:
                # Await next audio chunk with 5s timeout
                chunk = await asyncio.wait_for(ws_queue.get(), timeout=5.0)

                # Transcription MUST run in a background thread — never block the async event loop
                transcript = await asyncio.to_thread(transcriber.transcribe, chunk)

                if transcript:
                    # Clear queue backlog
                    while not ws_queue.empty():
                        try:
                            ws_queue.get_nowait()
                        except:
                            pass

                    outbound_queue.put_nowait({"type": "transcript", "text": transcript})

                    rag_context = ""
                    if config.RAG_ENABLED:
                        loop = asyncio.get_event_loop()
                        rag_context = await loop.run_in_executor(None, retriever.retrieve, transcript)

                    system_content = config.SYSTEM_PROMPT
                    if rag_context:
                        system_content += f"\n\n{rag_context}"
                        sources = list(dict.fromkeys(re.findall(r'\[Source: ([^\],]+)', rag_context)))
                        outbound_queue.put_nowait({"type": "rag_sources", "sources": sources})

                    messages = [
                        {"role": "system", "content": system_content},
                        {"role": "user", "content": transcript}
                    ]

                    try:
                        async for token in connection_llm_client.stream(messages):
                            outbound_queue.put_nowait({"type": "token", "text": token})
                        outbound_queue.put_nowait({"type": "end"})
                    except Exception as e:
                        print(f"LLM Stream Error: {e}", file=sys.stderr)
                        outbound_queue.put_nowait({"type": "error", "message": f"LLM Error: {str(e)}"})

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
        if ws_queue in active_ws_queues:
            active_ws_queues.remove(ws_queue)
        if outbound_queue in active_outbound_queues:
            active_outbound_queues.remove(outbound_queue)


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


if __name__ == "__main__":
    uvicorn.run("app:app", host=config.WS_HOST, port=config.WS_PORT, reload=False)
