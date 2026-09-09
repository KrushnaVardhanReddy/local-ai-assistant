import queue
import typing

# Global instances
transcriber: typing.Any = None
llm_client: typing.Any = None
listener: typing.Any = None
sync_queue: queue.Queue = queue.Queue()
candidate_context: str = ""
preferred_language: str = ""

# PTT Mode global toggle
PTT_MODE: bool = False

# Set of active per-connection asyncio.Queue objects for raw audio
active_ws_queues: set = set()

# Set of active per-connection asyncio.Queue objects for outbound JSON messages
active_outbound_queues: set = set()
