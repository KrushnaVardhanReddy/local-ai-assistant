#!/usr/bin/env python3
"""
Jules Batch Submitter — Local AI Assistant
Sends task prompts to Jules API to create async coding sessions → GitHub PRs.

Usage:
  python3 scripts/jules_submit.py --task P1-T1          # Submit a specific task
  python3 scripts/jules_submit.py --phase 1             # Submit all Phase 1 tasks
  python3 scripts/jules_submit.py --file prompts/tasks/P2_T1_audio_listener.txt
  python3 scripts/jules_submit.py --list                # List all available tasks
  python3 scripts/jules_submit.py --branch feature/dev  # Target a specific branch
"""

import json
import urllib.request
import sys
import os
import glob

# ──────────────────────────────────────────────────────────────────────────────
# Config — loads API key from .env.local or .env (never hardcode secrets)
# ──────────────────────────────────────────────────────────────────────────────

def _load_api_key():
    """Read JULES_API_KEY from environment, .env.local, or .env."""
    key = os.environ.get("JULES_API_KEY")
    if key:
        return key
    for envfile in [".env.local", ".env"]:
        repo_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        path = os.path.join(repo_root, envfile)
        if os.path.exists(path):
            with open(path) as f:
                for line in f:
                    line = line.strip()
                    if line.startswith("JULES_API_KEY="):
                        return line.split("=", 1)[1].strip()
    print("❌ JULES_API_KEY not found in environment, .env.local, or .env")
    sys.exit(1)

API_KEY  = _load_api_key()
API_URL  = "https://jules.googleapis.com/v1alpha/sessions"

# ── Repo root (one level up from scripts/) ────────────────────────────────────
REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# ── Update once the GitHub repo is created ───────────────────────────────────
REPO_SOURCE = "sources/github/KrushnaVardhanReddy/local-ai-assistant"

# Parse --branch from args
BRANCH = "main"
if "--branch" in sys.argv:
    idx = sys.argv.index("--branch")
    if idx + 1 < len(sys.argv):
        BRANCH = sys.argv[idx + 1]

# ──────────────────────────────────────────────────────────────────────────────
# Safety rules — prepended to every Jules prompt
# ──────────────────────────────────────────────────────────────────────────────

SAFETY_RULES = """
MANDATORY RULES — VIOLATION = REJECTED PR:
1. NEVER stub, mock, or TODO existing implementation code. Write real, working code only.
2. Commit message must start with "jules: " prefix.
3. Use clean OOP / module design. No spaghetti scripts.
4. NEVER hardcode secrets, API keys, or local paths — always read from environment variables.
5. All Python code must be compatible with Python 3.10+ and type-hinted where practical.

Project: Local AI Assistant
Tech stack:
  - Backend:  Python 3.10+, FastAPI, faster-whisper / Parakeet-TDT
  - LLM:      OpenAI-compatible /v1/chat/completions API
              Local:  Ollama (11434), LM Studio (1234), llama.cpp (8080) — no API key needed
              Cloud:  OpenAI, Groq, Gemini, Anthropic — API key via env var, Bearer auth header
              Config: LLM_PROVIDER env var selects provider; LLMClient.from_config() resolves all settings
  - Frontend: Svelte 5, Tauri 2.0, TypeScript
  - Bridge:   WebSocket (FastAPI → Svelte reactive store)
  - Target:   Dell XPS, 8GB VRAM, Linux/Windows desktop

Critical LLM rules:
- NEVER use Ollama-specific endpoints (/api/generate, /api/chat). Always use /v1/chat/completions.
- Cloud providers need `Authorization: Bearer <key>` header. Anthropic also needs `anthropic-version: 2023-06-01`.
- The LLMClient must work identically for all providers — only base_url, api_key, and headers differ.
- NEVER log or print API keys.
""".strip()

# ──────────────────────────────────────────────────────────────────────────────
# Task file conventions
#   prompts/tasks/P{phase}_T{task}_{slug}.txt  →  e.g. P1_T1_ollama_setup.txt
# ──────────────────────────────────────────────────────────────────────────────

TASKS_DIR = os.path.join(REPO_ROOT, "prompts", "tasks")

def _find_task_file(task_id: str) -> str:
    """Find a prompt file matching e.g. 'P1-T1' → prompts/tasks/**/P1_T1_*.txt"""
    normalized = task_id.replace("-", "_").upper()
    
    # Search recursively in subdirectories
    pattern = os.path.join(TASKS_DIR, "**", f"{normalized}_*.txt")
    matches = glob.glob(pattern, recursive=True)
    
    if not matches:
        # Fallback: exact filename match in any subdirectory
        exact_pattern = os.path.join(TASKS_DIR, "**", f"{normalized}.txt")
        exact_matches = glob.glob(exact_pattern, recursive=True)
        if exact_matches:
            return exact_matches[0]
            
        # Try finding if user passed the exact filename minus extension
        file_pattern = os.path.join(TASKS_DIR, "**", f"{task_id}.txt")
        file_matches = glob.glob(file_pattern, recursive=True)
        if file_matches:
            return file_matches[0]
            
        print(f"❌ No prompt file found for task '{task_id}' in {TASKS_DIR}/ (including subfolders)")
        print(f"   Expected pattern: {normalized}_<slug>.txt")
        sys.exit(1)
        
    if len(matches) > 1:
        print(f"⚠️  Multiple files found for '{task_id}': {matches}")
        print(f"   Using: {matches[0]}")
    return matches[0]

def _find_phase_files(phase_num: int) -> list:
    """Find all prompt files for a given phase, sorted by task number."""
    pattern = os.path.join(TASKS_DIR, "**", f"P{phase_num}_T*.txt")
    matches = sorted(glob.glob(pattern, recursive=True))
    if not matches:
        print(f"❌ No prompt files found for Phase {phase_num} in {TASKS_DIR}/ (including subfolders)")
        sys.exit(1)
    return matches

# ──────────────────────────────────────────────────────────────────────────────
# Submission logic
# ──────────────────────────────────────────────────────────────────────────────

def submit_prompt(full_prompt: str, task_name: str = "Task"):
    payload = json.dumps({
        "prompt": full_prompt,
        "sourceContext": {
            "source": REPO_SOURCE,
            "githubRepoContext": {
                "startingBranch": BRANCH
            }
        }
    }).encode()

    req = urllib.request.Request(
        API_URL,
        data=payload,
        headers={
            "Content-Type": "application/json",
            "x-goog-api-key": API_KEY
        },
        method="POST"
    )

    print(f"🚀 Submitting: {task_name} → branch: {BRANCH}")
    try:
        with urllib.request.urlopen(req) as resp:
            result = json.loads(resp.read())
            session_id = result.get("name", "unknown").split("/")[-1]
            print(f"✅ Session created: {session_id}")
            print(f"   View at: https://jules.google.com/session/{session_id}")
            return session_id
    except urllib.error.HTTPError as e:
        print(f"❌ HTTP {e.code}: {e.read().decode()}")
        sys.exit(1)

def submit_file(filepath: str, label: str = None):
    if not os.path.exists(filepath):
        print(f"❌ File not found: {filepath}")
        sys.exit(1)
    with open(filepath) as f:
        prompt_content = f.read()
    full_prompt = SAFETY_RULES + "\n\n---\n\n" + prompt_content
    name = label or os.path.basename(filepath)
    submit_prompt(full_prompt, task_name=name)

def list_tasks():
    """Print all available task prompt files."""
    if not os.path.isdir(TASKS_DIR):
        print(f"❌ Tasks directory not found: {TASKS_DIR}")
        sys.exit(1)
    files = sorted(glob.glob(os.path.join(TASKS_DIR, "*.txt")))
    if not files:
        print("⚠️  No task prompt files found in prompts/tasks/")
        return
    print(f"\n📋 Available tasks ({len(files)} total):\n")
    for f in files:
        basename = os.path.basename(f)
        print(f"   {basename}")
    print()

# ──────────────────────────────────────────────────────────────────────────────
# CLI entry point
# ──────────────────────────────────────────────────────────────────────────────

def main():
    args = sys.argv[1:]

    if not args or "--help" in args or "-h" in args:
        print(__doc__)
        sys.exit(0)

    if "--list" in args:
        list_tasks()
        sys.exit(0)

    if "--file" in args:
        idx = args.index("--file")
        if idx + 1 >= len(args):
            print("❌ Please specify a file path after --file.")
            sys.exit(1)
        submit_file(args[idx + 1])
        sys.exit(0)

    if "--task" in args:
        idx = args.index("--task")
        if idx + 1 >= len(args):
            print("❌ Please specify a task ID after --task (e.g. P1-T1).")
            sys.exit(1)
        task_id = args[idx + 1]
        filepath = _find_task_file(task_id)
        submit_file(filepath, label=task_id)
        sys.exit(0)

    if "--phase" in args:
        idx = args.index("--phase")
        if idx + 1 >= len(args):
            print("❌ Please specify a phase number after --phase.")
            sys.exit(1)
        phase_num = int(args[idx + 1])
        files = _find_phase_files(phase_num)
        print(f"📅 Submitting {len(files)} task(s) for Phase {phase_num}...")
        for filepath in files:
            submit_file(filepath, label=os.path.basename(filepath))
        sys.exit(0)

    print(__doc__)

if __name__ == "__main__":
    main()
