import asyncio
from backend.llm_client import LLMClient

async def main():
    client = await LLMClient.from_config()
    print(f"Using model: {client.model}")
    messages = [{"role": "user", "content": "Hello! Please reply in one sentence."}]
    async for token in client.stream(messages):
        pass

asyncio.run(main())
