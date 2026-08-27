import sys
import os
import urllib.request
import urllib.error
import json
from config import config

LOCAL_PROVIDERS = {
    "ollama": "http://localhost:11434/v1",
    "lmstudio": "http://localhost:1234/v1",
    "llamacpp": "http://localhost:8080/v1",
}

def probe(url: str, timeout: float = 1.0) -> bool:
    try:
        req = urllib.request.Request(f"{url}/models")
        with urllib.request.urlopen(req, timeout=timeout) as response:
            if response.status == 200:
                return True
    except Exception:
        pass
    return False

def detect_backend() -> tuple[str, str]:
    for provider, base_url in LOCAL_PROVIDERS.items():
        if probe(base_url):
            return provider, base_url
    print("❌ No local LLM backend detected.")
    sys.exit(1)

def test_completion(base_url: str, model: str, api_key: str = "", provider: str = "") -> bool:
    payload = {
        "model": model,
        "messages": [{"role": "user", "content": "Say OK"}],
        "max_tokens": 5
    }
    data = json.dumps(payload).encode('utf-8')
    req = urllib.request.Request(f"{base_url}/chat/completions", data=data, method="POST")
    req.add_header('Content-Type', 'application/json')
    if api_key:
        req.add_header('Authorization', f'Bearer {api_key}')
    if provider == "anthropic":
        req.add_header('anthropic-version', '2023-06-01')

    try:
        # We give the completion test a bit more time than the probe to generate tokens
        with urllib.request.urlopen(req, timeout=10.0) as response:
            if response.status == 200:
                resp_data = json.loads(response.read().decode('utf-8'))
                if "choices" in resp_data and len(resp_data["choices"]) > 0:
                    content = resp_data["choices"][0].get("message", {}).get("content", "")
                    if content:
                        return True
    except Exception as e:
        print(f"❌ Completion test failed: {e}")
        pass
    return False

def main():
    if config.LLM_PROVIDER == "auto":
        provider, base_url = detect_backend()
        api_key = ""
        print(f"✅ Auto-detected backend: {provider} at {base_url}")
    else:
        resolved = config.resolved_llm()
        provider = resolved["provider"]
        base_url = resolved["base_url"]
        api_key = resolved["api_key"]
        print(f"✅ Using configured backend: {provider} at {base_url}")

    model = config.LLM_MODEL
    if test_completion(base_url, model, api_key, provider):
        print(f"✅ Completion successful.")
        print(f"Active base URL: {base_url}")
        sys.exit(0)
    else:
        sys.exit(1)

if __name__ == "__main__":
    main()
