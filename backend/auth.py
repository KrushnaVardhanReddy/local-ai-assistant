import jwt
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
