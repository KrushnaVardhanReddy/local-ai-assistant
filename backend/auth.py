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

async def process_referral_reward(stripe_customer_id: str) -> None:
    """Checks if the user has an associated referral, adds a $5 credit (1 payg_session) to the referrer, and marks the referral as converted."""
    url = f"{config.SUPABASE_URL.rstrip('/')}/rest/v1/profiles"
    headers = {
        "apikey": config.SUPABASE_SERVICE_KEY,
        "Authorization": f"Bearer {config.SUPABASE_SERVICE_KEY}",
        "Content-Type": "application/json"
    }

    async with httpx.AsyncClient() as client:
        # First, find the user_id for the given stripe_customer_id
        profile_params = {"stripe_customer_id": f"eq.{stripe_customer_id}", "select": "id"}
        profile_resp = await client.get(url, headers=headers, params=profile_params)
        profile_resp.raise_for_status()
        profile_data = profile_resp.json()

        if not profile_data or len(profile_data) == 0:
            return

        user_id = profile_data[0].get("id")

        # Check if this user was referred
        referrals_url = f"{config.SUPABASE_URL.rstrip('/')}/rest/v1/referrals"
        referral_params = {
            "referee_id": f"eq.{user_id}",
            "status": "neq.converted",
            "select": "id,referrer_id"
        }

        referral_resp = await client.get(referrals_url, headers=headers, params=referral_params)
        referral_resp.raise_for_status()
        referral_data = referral_resp.json()

        if not referral_data or len(referral_data) == 0:
            return

        referral_id = referral_data[0].get("id")
        referrer_id = referral_data[0].get("referrer_id")

        # Add $5 credit to the referrer's account (1 payg_session)
        # Using grant_payg_session which already increments payg_sessions by 1
        await grant_payg_session(referrer_id)

        # Mark the referral as converted
        update_referral_payload = {
            "status": "converted",
            "converted_at": datetime.datetime.utcnow().isoformat()
        }
        update_referral_params = {"id": f"eq.{referral_id}"}

        patch_referral_resp = await client.patch(referrals_url, headers=headers, params=update_referral_params, json=update_referral_payload)
        patch_referral_resp.raise_for_status()
