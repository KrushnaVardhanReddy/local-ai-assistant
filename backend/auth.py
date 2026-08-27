import jwt
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
