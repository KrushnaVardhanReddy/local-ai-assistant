import jwt
import datetime
import httpx
from config import config

class AuthError(Exception):
    pass

def verify_supabase_jwt(token: str) -> dict:
    try:
        payload = jwt.decode(
            token,
            config.SUPABASE_JWT_SECRET,
            algorithms=["HS256"],
            audience="authenticated"
        )
        return payload
    except (jwt.ExpiredSignatureError, jwt.InvalidTokenError, jwt.InvalidAudienceError) as e:
        raise AuthError(f"Invalid token: {str(e)}")

def get_user_id(token: str) -> str:
    payload = verify_supabase_jwt(token)
    user_id = payload.get("sub")
    if not user_id:
        raise AuthError("Token missing 'sub' claim")
    return user_id

async def update_user_plan(user_id: str, new_plan: str) -> None:
    """Updates the user's plan in Supabase."""
    url = f"{config.SUPABASE_URL.rstrip('/')}/rest/v1/profiles"
    headers = {
        "apikey": config.SUPABASE_SERVICE_KEY,
        "Authorization": f"Bearer {config.SUPABASE_SERVICE_KEY}",
        "Content-Type": "application/json"
    }
    params = {"id": f"eq.{user_id}"}
    payload = {"plan": new_plan}

    async with httpx.AsyncClient() as client:
        resp = await client.patch(url, headers=headers, params=params, json=payload)
        resp.raise_for_status()

async def grant_payg_session(user_id: str) -> None:
    """Grants a pay-as-you-go session by incrementing the payg_sessions counter."""
    url = f"{config.SUPABASE_URL.rstrip('/')}/rest/v1/profiles"
    headers = {
        "apikey": config.SUPABASE_SERVICE_KEY,
        "Authorization": f"Bearer {config.SUPABASE_SERVICE_KEY}",
        "Content-Type": "application/json"
    }
    params = {"id": f"eq.{user_id}", "select": "payg_sessions"}

    async with httpx.AsyncClient() as client:
        # Get current sessions
        resp = await client.get(url, headers=headers, params=params)
        resp.raise_for_status()
        data = resp.json()

        if data and len(data) > 0:
            current_sessions = data[0].get("payg_sessions") or 0

            # Increment and update
            update_payload = {"payg_sessions": current_sessions + 1}
            update_params = {"id": f"eq.{user_id}"}
            patch_resp = await client.patch(url, headers=headers, params=update_params, json=update_payload)
            patch_resp.raise_for_status()

def create_payg_session_token(user_id: str) -> str:
    """Issues a 90-minute session token for payg users."""
    payload = {
        "sub": user_id,
        "type": "payg_session",
        "exp": datetime.datetime.utcnow() + datetime.timedelta(minutes=90)
    }
    return jwt.encode(payload, config.SUPABASE_JWT_SECRET, algorithm="HS256")

def verify_payg_session_token(token: str) -> dict:
    """Verifies the payg session token."""
    try:
        payload = jwt.decode(
            token,
            config.SUPABASE_JWT_SECRET,
            algorithms=["HS256"]
        )
        return payload
    except (jwt.ExpiredSignatureError, jwt.InvalidTokenError) as e:
        raise AuthError(f"Session token invalid or expired: {str(e)}")

async def get_user_plan(user_id: str) -> str:
    """Fetches the user's plan from Supabase."""
    url = f"{config.SUPABASE_URL.rstrip('/')}/rest/v1/profiles"
    headers = {
        "apikey": config.SUPABASE_SERVICE_KEY,
        "Authorization": f"Bearer {config.SUPABASE_SERVICE_KEY}",
        "Content-Type": "application/json"
    }
    params = {"id": f"eq.{user_id}", "select": "plan"}

    async with httpx.AsyncClient() as client:
        resp = await client.get(url, headers=headers, params=params)
        resp.raise_for_status()
        data = resp.json()
        if data and len(data) > 0:
            return data[0].get("plan", "unknown")
        return "unknown"
