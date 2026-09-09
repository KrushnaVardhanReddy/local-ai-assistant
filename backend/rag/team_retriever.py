import os
from typing import List, Dict, Any
from supabase import create_client, Client
from config import config

class TeamRetriever:
    def __init__(self, org_id: str | None, query_embedding: List[float]):
        self.org_id = org_id
        self.query_embedding = query_embedding
        self.supabase: Client | None = None
        if config.SUPABASE_URL and config.SUPABASE_SERVICE_KEY:
            self.supabase = create_client(config.SUPABASE_URL, config.SUPABASE_SERVICE_KEY)

    def retrieve(self, top_k: int = 4) -> List[tuple[str, str, str]]:
        if not self.org_id or not self.supabase:
            return []

        try:
            response = self.supabase.rpc(
                "match_team_documents",
                {
                    "query_embedding": self.query_embedding,
                    "match_threshold": 0.0,
                    "match_count": top_k,
                    "p_org_id": self.org_id
                }
            ).execute()

            data = response.data
            if not data:
                return []

            kept = []
            for row in data:
                content = row.get("content", "")
                metadata = row.get("metadata", {})
                source = metadata.get("source_file", "team_document")
                chunk_index = metadata.get("chunk_index", "?")
                kept.append((source, str(chunk_index), content))

            return kept
        except Exception as e:
            print(f"Error querying team documents: {e}")
            return []
