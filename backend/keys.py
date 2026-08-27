import hashlib
import httpx
from cryptography.fernet import Fernet
from config import config

class KeyStore:
    def __init__(self, encryption_key: str, supabase_url: str, service_key: str):
        if not encryption_key:
            raise ValueError("Encryption key cannot be empty")
        self._fernet = Fernet(encryption_key.encode())
        self.supabase_url = supabase_url.rstrip('/')
        self.service_key = service_key

    def _get_headers(self) -> dict:
        return {
            "apikey": self.service_key,
            "Authorization": f"Bearer {self.service_key}",
            "Content-Type": "application/json"
        }

    async def get_key(self, user_id: str, provider: str) -> str | None:
        url = f"{self.supabase_url}/rest/v1/user_api_keys"
        params = {
            "user_id": f"eq.{user_id}",
            "provider": f"eq.{provider}",
            "select": "key_encrypted"
        }

        async with httpx.AsyncClient() as client:
            resp = await client.get(url, headers=self._get_headers(), params=params)

        if resp.status_code == 200:
            data = resp.json()
            if data and len(data) > 0:
                key_encrypted = data[0].get("key_encrypted")
                if key_encrypted:
                    decrypted_key = self._fernet.decrypt(key_encrypted.encode()).decode()
                    return decrypted_key
        return None

    async def save_key(self, user_id: str, provider: str, api_key: str) -> None:
        encrypted = self._fernet.encrypt(api_key.encode()).decode()
        key_hash = hashlib.sha256(api_key.encode()).hexdigest()[:12]

        url = f"{self.supabase_url}/rest/v1/user_api_keys"
        payload = {
            "user_id": user_id,
            "provider": provider,
            "key_encrypted": encrypted,
            "key_hash": key_hash
        }
        headers = self._get_headers()
        headers["Prefer"] = "resolution=merge-duplicates"

        async with httpx.AsyncClient() as client:
            resp = await client.post(url, headers=headers, json=payload)
            resp.raise_for_status()

    async def verify_user_profile(self, user_id: str, machine_id: str) -> bool:
        url = f"{self.supabase_url}/rest/v1/profiles"
        params = {"id": f"eq.{user_id}", "select": "plan,machine_id"}
        async with httpx.AsyncClient() as client:
            resp = await client.get(url, headers=self._get_headers(), params=params)

        if resp.status_code != 200 or not resp.json():
            raise Exception("Profile not found")

        profile = resp.json()[0]
        if profile.get("plan") != "lifetime":
            raise Exception("License required")

        db_machine_id = profile.get("machine_id")
        if not db_machine_id:
            patch_payload = {"machine_id": machine_id}
            async with httpx.AsyncClient() as client:
                await client.patch(url, headers=self._get_headers(), params={"id": f"eq.{user_id}"}, json=patch_payload)
        elif db_machine_id != machine_id:
            raise Exception("Hardware ID mismatch. Please reset your machine ID in the dashboard.")

        return True

    async def delete_key(self, user_id: str, provider: str) -> None:
        url = f"{self.supabase_url}/rest/v1/user_api_keys"
        params = {
            "user_id": f"eq.{user_id}",
            "provider": f"eq.{provider}"
        }

        async with httpx.AsyncClient() as client:
            resp = await client.delete(url, headers=self._get_headers(), params=params)
            resp.raise_for_status()

key_store = KeyStore(config.ENCRYPTION_KEY, config.SUPABASE_URL, config.SUPABASE_SERVICE_KEY) if config.ENCRYPTION_KEY else None
