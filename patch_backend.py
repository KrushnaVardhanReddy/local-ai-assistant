import re

with open("backend/app.py", "r") as f:
    content = f.read()

endpoint_code = """
@app.get("/api/status")
async def get_system_status():
    return {
        "llm_provider": config.LLM_PROVIDER,
        "llm_model": config.LLM_MODEL,
        "stt_provider": config.STT_PROVIDER,
        "stt_model": config.STT_MODEL,
        "local_stt_engine": config.LOCAL_STT_ENGINE
    }

@app.get("/api/resume/context")
"""

content = content.replace("@app.get(\"/api/resume/context\")", endpoint_code)

with open("backend/app.py", "w") as f:
    f.write(content)
