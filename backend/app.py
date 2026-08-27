import asyncio
import queue
import threading
import traceback
import sys

from fastapi import FastAPI, WebSocket, WebSocketDisconnect
from fastapi.responses import JSONResponse
import uvicorn

from config import config
from audio_listener import AudioListener
from transcriber import Transcriber
from llm_client import LLMClient


from contextlib import asynccontextmanager

@asynccontextmanager
async def lifespan(app: FastAPI):
    await startup_event()
    yield
    await shutdown_event()

app = FastAPI(lifespan=lifespan)


# Global instances
transcriber = None
llm_client = None
listener = None
sync_queue = queue.Queue()

# Set of active per-connection asyncio.Queue objects
active_ws_queues = set()

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

    if config.SUPABASE_JWT_SECRET:
        await websocket.close(code=1008, reason="Auth not yet implemented — set P6-T4 task")
        return

    # Create a per-connection asyncio.Queue
    ws_queue = asyncio.Queue()
    active_ws_queues.add(ws_queue)

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

                    await websocket.send_json({"type": "transcript", "text": transcript})

                    messages = [{"role": "user", "content": transcript}]

                    try:
                        async for token in llm_client.stream(messages):
                            await websocket.send_json({"type": "token", "text": token})
                        await websocket.send_json({"type": "end"})
                    except Exception as e:
                        print(f"LLM Stream Error: {e}", file=sys.stderr)
                        await websocket.send_json({"type": "error", "message": f"LLM Error: {str(e)}"})

            except asyncio.TimeoutError:
                # Keep-alive ping
                # Preventing the WebSocket from closing on long silences
                try:
                    await websocket.send_json({"type": "ping"})
                except:
                    pass

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
        if ws_queue in active_ws_queues:
            active_ws_queues.remove(ws_queue)


@app.get("/health")
async def health_check():
    return {
        "status": "ok",
        "llm_provider": config.LLM_PROVIDER,
        "auth_enabled": bool(config.SUPABASE_JWT_SECRET)
    }


if __name__ == "__main__":
    uvicorn.run("app:app", host=config.WS_HOST, port=config.WS_PORT, reload=False)
