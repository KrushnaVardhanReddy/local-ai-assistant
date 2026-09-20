import sys

content = open("wiki/architecture.md").read()

if "llm-proxy" not in content:
    content = content.replace("All sensitive tokens and activation secrets are stored natively on the user's OS Keychain using `zalando/go-keyring`.", "All sensitive tokens and activation secrets are stored natively on the user's OS Keychain using `zalando/go-keyring`. During the 15-minute free demo, the app uses a Supabase Edge Function (`llm-proxy`) to securely proxy LLM API requests and inject our server-side API key without exposing it to the client binary.")

open("wiki/architecture.md", "w").write(content)
