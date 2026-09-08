import asyncio
import json
import aiohttp
import sys
from typing import AsyncGenerator

from config import config

class LLMClient:
    def __init__(self, base_url: str, model: str, api_key: str = "", provider: str = ""):
        self.base_url = base_url.rstrip("/") if base_url else ""
        self.model = model
        self.api_key = api_key
        self.provider = provider

    def _build_headers(self) -> dict:
        headers = {"Content-Type": "application/json"}
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"
        if self.provider.lower() == "anthropic":
            headers["anthropic-version"] = "2023-06-01"
        return headers

    async def stream(self, messages: list[dict], system_prompt: str = "") -> AsyncGenerator[str, None]:
        msgs = messages.copy()
        if system_prompt:
            msgs.insert(0, {"role": "system", "content": system_prompt})

        payload = {
            "model": self.model,
            "messages": msgs,
            "stream": True
        }

        try:
            async with aiohttp.ClientSession() as session:
                print(f"[DEBUG LLM] Sending POST to {self.base_url}/chat/completions", file=sys.stderr)
                async with session.post(f"{self.base_url}/chat/completions", headers=self._build_headers(), json=payload) as resp:
                    print(f"[DEBUG LLM] Got response status: {resp.status}", file=sys.stderr)
                    if resp.status != 200:
                        if resp.status in (401, 403):
                            yield f"[Error: Invalid API key for {self.provider}]"
                        elif resp.status == 429:
                            yield "[Error: Rate limit hit — please retry shortly]"
                        else:
                            error_text = await resp.text()
                            yield f"[Error: LLM returned status {resp.status} - {error_text}]"
                        return

                    async for line in resp.content:
                        decoded_line = line.decode('utf-8').strip()
                        print(f"[DEBUG LLM] Raw line: {decoded_line}", file=sys.stderr)
                        if not decoded_line:
                            continue

                        if decoded_line.startswith("data: "):
                            data_str = decoded_line[len("data: "):]
                            if data_str == "[DONE]":
                                break

                            try:
                                chunk = json.loads(data_str)
                                choices = chunk.get("choices", [])
                                if choices:
                                    delta = choices[0].get("delta", {})
                                    
                                    # Handle DeepSeek/reasoning models
                                    reasoning = delta.get("reasoning")
                                    if reasoning:
                                        yield reasoning
                                        
                                    content = delta.get("content")
                                    if content is not None:
                                        yield content
                                else:
                                    print(f"[DEBUG LLM] Empty choices in chunk: {chunk}", file=sys.stderr)
                            except json.JSONDecodeError:
                                pass
                        else:
                            print(f"[DEBUG LLM] Ignored non-data line: {decoded_line}", file=sys.stderr)
        except aiohttp.ClientConnectorError:
            yield f"[Error: Cannot connect to LLM backend at {self.base_url}]"

    async def generate_scorecard(self, session_data: dict) -> dict:
        """
        Generate a structured interview scorecard from session turn data.

        Args:
            session_data: Output of SessionManager.export() — contains turns list
                          with transcripts and responses.

        Returns:
            dict with keys: overall_score, summary, turns (list of per-turn evals)
        """
        turns = session_data.get("turns", [])
        if not turns:
            return {"error": "No session data to score"}

        # Build a compact transcript for the LLM
        transcript_lines = []
        for t in turns:
            transcript_lines.append(f"Q{t['turn'] + 1}: {t['transcript']}")
            transcript_lines.append(f"A{t['turn'] + 1}: {t['response']}")
        transcript_text = "\n".join(transcript_lines)

        scorecard_prompt = f"""You are an expert technical interview evaluator. Analyze the following interview transcript and return a JSON scorecard.

TRANSCRIPT:
{transcript_text}

Return ONLY valid JSON with this exact structure (no markdown, no explanation):
{{
  "overall_score": <1-10 integer>,
  "overall_summary": "<2-3 sentence overall assessment>",
  "strengths": ["<strength 1>", "<strength 2>"],
  "gaps": ["<gap 1>", "<gap 2>"],
  "turns": [
    {{
      "turn": 0,
      "score": <1-5 integer>,
      "verdict": "<Good|Partial|Incomplete|Off-topic>",
      "what_was_good": "<one sentence>",
      "what_was_missing": "<one sentence or null>",
      "suggested_addition": "<one concise sentence the candidate should have said, or null>"
    }}
  ]
}}"""

        messages = [
            {"role": "system", "content": "You are a precise JSON-only technical interview evaluator."},
            {"role": "user", "content": scorecard_prompt}
        ]

        import httpx

        headers = self._build_headers()
        payload = {
            "model": self.model,
            "messages": messages,
            "stream": False,
            "temperature": 0.3,
            "max_tokens": 2000,
        }

        try:
            async with httpx.AsyncClient(timeout=30) as client:
                resp = await client.post(
                    f"{self.base_url}/chat/completions",
                    headers=headers,
                    json=payload,
                )
                resp.raise_for_status()
                content = resp.json()["choices"][0]["message"]["content"].strip()
                if content.startswith("```"):
                    content = content.split("```", 1)[1]
                    if content.startswith("json\n"):
                        content = content[5:]
                    elif content.startswith("json"):
                        content = content[4:]

                    if content.endswith("```"):
                        content = content[:-3].strip()
                return json.loads(content)
        except Exception as e:
            return {"error": f"Scorecard generation failed: {str(e)}"}

    async def health_check(self) -> bool:
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(f"{self.base_url}/models", headers=self._build_headers()) as resp:
                    if resp.status == 200:
                        text = await resp.text()
                        if self.model in text:
                            return True
            return False
        except Exception:
            return False

    async def chat_vision(self, image_base64: str, prompt: str) -> AsyncGenerator[str, None]:
        if not image_base64.startswith("data:image/"):
            image_base64 = f"data:image/jpeg;base64,{image_base64}"

        messages = [{
            "role": "user",
            "content": [
                {"type": "text", "text": prompt},
                {"type": "image_url", "image_url": {"url": image_base64}}
            ]
        }]

        async for token in self.stream(messages):
            yield token

    @classmethod
    async def from_config(cls, is_vision: bool = False, api_key_override: str = None) -> "LLMClient":
        llm_config = config.resolved_llm()
        base_url = llm_config.get("base_url", "")
        api_key = llm_config.get("api_key", "")
        if api_key_override:
            api_key = api_key_override
        provider = llm_config.get("provider", "")
        model = config.VISION_MODEL if is_vision else config.LLM_MODEL

        if provider == "auto":
            candidates = {
                "ollama": "http://localhost:11434/v1",
                "lmstudio": "http://localhost:1234/v1",
                "llamacpp": "http://localhost:8080/v1"
            }
            for cand, cand_url in candidates.items():
                client = cls(cand_url, model, "", cand)
                try:
                    is_healthy = await asyncio.wait_for(client.health_check(), timeout=1.0)
                    if is_healthy:
                        base_url = cand_url
                        provider = cand
                        break
                except (asyncio.TimeoutError, Exception):
                    continue

        return cls(base_url, model, api_key, provider)


async def test():
    client = await LLMClient.from_config()
    messages = [{"role": "user", "content": "What is 2+2? Answer in one word."}]
    async for token in client.stream(messages):
        print(token, end="", flush=True)
    print()


if __name__ == "__main__":
    asyncio.run(test())
