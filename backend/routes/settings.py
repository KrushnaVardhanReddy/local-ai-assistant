from fastapi import APIRouter, Depends, HTTPException, Request, Response
from pydantic import BaseModel
from config import config, PROVIDER_CONFIG
from keys import key_store
from auth import get_user_id, AuthError
from audio_listener import get_audio_devices
import state

router = APIRouter()

class LanguagePreference(BaseModel):
    language: str

@router.post("/api/language")
async def set_language(body: LanguagePreference):
    state.preferred_language = body.language
    return {"status": "ok", "language": state.preferred_language}

@router.get("/api/language")
async def get_language():
    return {"language": state.preferred_language}

class InterviewLanguage(BaseModel):
    language: str

@router.post("/api/interview_language")
async def set_interview_language(body: InterviewLanguage):
    old_override = config.LANGUAGE_OVERRIDE
    config.LANGUAGE_OVERRIDE = body.language

    # Remove previous override string if present
    override_str_format = " You must respond entirely in the {lang} language, except for code snippets."
    if old_override != "auto":
        old_str = override_str_format.format(lang=old_override)
        config.SYSTEM_PROMPT = config.SYSTEM_PROMPT.replace(old_str, "")

    if config.LANGUAGE_OVERRIDE != "auto":
        config.SYSTEM_PROMPT += override_str_format.format(lang=config.LANGUAGE_OVERRIDE)

    return {"status": "ok", "language": config.LANGUAGE_OVERRIDE}

@router.get("/api/interview_language")
async def get_interview_language():
    return {"language": config.LANGUAGE_OVERRIDE}

class JobDescriptionModel(BaseModel):
    text: str

@router.post("/config/job-description")
async def update_job_description(body: JobDescriptionModel):
    config.JOB_DESCRIPTION = body.text
    return {"status": "success"}

class PromptModel(BaseModel):
    prompt: str

@router.get("/api/system_prompt")
async def get_system_prompt():
    return {"prompt": config.SYSTEM_PROMPT}

@router.post("/api/system_prompt")
async def set_system_prompt(body: PromptModel):
    config.SYSTEM_PROMPT = body.prompt
    return {"status": "success"}

class DeviceModel(BaseModel):
    is_loopback: bool
    device_id: int | None

@router.get("/api/audio/devices")
async def get_audio_devices_endpoint():
    return get_audio_devices()

@router.post("/api/audio/device")
async def set_audio_device(body: DeviceModel):
    config.AUDIO_LOOPBACK = body.is_loopback
    config.AUDIO_DEVICE = body.device_id

    # If the listener is currently running, we need to restart it
    if state.listener:
        state.listener.stop()
        state.listener.device = body.device_id
        state.listener.is_loopback = body.is_loopback
        state.listener.start()

    return {"status": "success", "device_id": body.device_id, "is_loopback": body.is_loopback}


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

@router.post("/api/keys/{provider}")
async def save_api_key(provider: str, body: KeyModel, user_id: str = Depends(get_current_user_id)):
    if not key_store:
        raise HTTPException(status_code=500, detail="Key store not configured")
    await key_store.save_key(user_id, provider, body.api_key)
    return {"message": "Key saved successfully"}

@router.delete("/api/keys/{provider}")
async def delete_api_key(provider: str, user_id: str = Depends(get_current_user_id)):
    if not key_store:
        raise HTTPException(status_code=500, detail="Key store not configured")
    await key_store.delete_key(user_id, provider)
    return Response(status_code=204)

@router.post("/api/keys/{provider}/test")
async def test_api_key(provider: str, body: KeyModel, user_id: str = Depends(get_current_user_id)):
    from llm_client import LLMClient
    base_url = PROVIDER_CONFIG.get(provider, {}).get("url", "")
    temp_client = LLMClient(base_url=base_url, model=config.LLM_MODEL, api_key=body.api_key, provider=provider)
    is_ok = await temp_client.health_check()
    return {"ok": is_ok}

@router.get("/health")
async def health_check():
    return {
        "status": "ok",
        "llm_provider": config.LLM_PROVIDER,
        "auth_enabled": bool(config.SUPABASE_JWT_SECRET)
    }

@router.get("/rag/status")
async def rag_status():
    return {"enabled": config.RAG_ENABLED}

@router.post("/rag/toggle")
async def rag_toggle():
    config.RAG_ENABLED = not config.RAG_ENABLED
    return {"enabled": config.RAG_ENABLED}

@router.get("/web_search/status")
async def web_search_status():
    return {"enabled": config.WEB_SEARCH_ENABLED}

@router.post("/web_search/toggle")
async def web_search_toggle():
    config.WEB_SEARCH_ENABLED = not config.WEB_SEARCH_ENABLED
    return {"enabled": config.WEB_SEARCH_ENABLED}

class InternalPlanUpdateModel(BaseModel):
    user_id: str
    plan: str

@router.post("/api/internal/billing/update_plan")
async def internal_update_plan(body: InternalPlanUpdateModel, request: Request):
    from auth import update_user_plan, grant_payg_session
    secret = request.headers.get("x-internal-secret")
    if not secret or secret != config.SUPABASE_SERVICE_KEY:
        raise HTTPException(status_code=403, detail="Forbidden")

    try:
        if body.plan == "payg":
            await grant_payg_session(body.user_id)
            await update_user_plan(body.user_id, body.plan)
        else:
            await update_user_plan(body.user_id, body.plan)
        return {"status": "success"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
