with open("backend/app.py", "r") as f:
    content = f.read()
content = content.replace('print(f"[DEBUG WS]', '# print(f"[DEBUG WS]')
content = content.replace('print(f"[DEBUG LLM]', '# print(f"[DEBUG LLM]')
with open("backend/app.py", "w") as f:
    f.write(content)

with open("backend/llm_client.py", "r") as f:
    content = f.read()
content = content.replace('print(f"[DEBUG WS]', '# print(f"[DEBUG WS]')
content = content.replace('print(f"[DEBUG LLM]', '# print(f"[DEBUG LLM]')
with open("backend/llm_client.py", "w") as f:
    f.write(content)

with open("backend/routes/ws.py", "r") as f:
    content = f.read()
content = content.replace('print(f"[DEBUG WS]', '# print(f"[DEBUG WS]')
content = content.replace('print(f"[DEBUG LLM]', '# print(f"[DEBUG LLM]')
content = content.replace('print(f"[DEBUG] Calling LLM', '# print(f"[DEBUG] Calling LLM')
content = content.replace('print(f"[DEBUG] Finished calling LLM', '# print(f"[DEBUG] Finished calling LLM')
with open("backend/routes/ws.py", "w") as f:
    f.write(content)
