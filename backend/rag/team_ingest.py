import os
import pathlib
import shutil
from fastapi import APIRouter, UploadFile, File, Depends, HTTPException, Request
from supabase import create_client, Client
from config import config
from rag.ingestor import extract_text, chunk_text, get_embedder
from auth import get_user_id, AuthError

router = APIRouter()

def get_current_user_id(request: Request) -> str | None:
    auth_header = request.headers.get("Authorization")
    if not auth_header or not auth_header.startswith("Bearer "):
        return None
    token = auth_header.split(" ")[1]
    try:
        return get_user_id(token)
    except AuthError:
        return None

def get_supabase_client() -> Client:
    if not config.SUPABASE_URL or not config.SUPABASE_SERVICE_KEY:
        raise HTTPException(status_code=500, detail="Supabase not configured")
    return create_client(config.SUPABASE_URL, config.SUPABASE_SERVICE_KEY)

async def get_user_org(user_id: str) -> str | None:
    from auth import get_user_org as auth_get_user_org
    return await auth_get_user_org(user_id)

@router.post("/team/ingest")
async def ingest_team_document(
    file: UploadFile = File(...),
    user_id: str = Depends(get_current_user_id)
):
    if not user_id:
        raise HTTPException(status_code=401, detail="Unauthorized")

    org_id = await get_user_org(user_id)
    if not org_id:
        raise HTTPException(status_code=403, detail="User is not part of an enterprise organization")

    upload_dir = pathlib.Path(config.UPLOAD_DIR)
    safe_filename = pathlib.Path(file.filename).name
    file_path = upload_dir / safe_filename

    try:
        # Save file temporarily
        with open(file_path, "wb") as buffer:
            shutil.copyfileobj(file.file, buffer)

        size_bytes = file_path.stat().st_size

        # Extract text and chunk
        text = extract_text(file_path)
        if not text.strip():
            raise ValueError("Document appears to be empty or unreadable")

        chunks = chunk_text(text)
        embedder = get_embedder()
        embeddings = embedder.encode(chunks, batch_size=32, show_progress_bar=False).tolist()

        supabase = get_supabase_client()

        # Insert document chunks
        docs_data = []
        for i, (chunk, embedding) in enumerate(zip(chunks, embeddings)):
            docs_data.append({
                "org_id": org_id,
                "content": chunk,
                "embedding": embedding,
                "metadata": {
                    "source_file": safe_filename,
                    "chunk_index": i,
                    "file_type": file_path.suffix.lower()
                }
            })

        if docs_data:
            supabase.table("team_documents").insert(docs_data).execute()

        # Insert metadata
        meta_data = {
            "org_id": org_id,
            "filename": safe_filename,
            "uploaded_by": user_id,
            "size_bytes": size_bytes
        }
        supabase.table("team_documents_meta").insert(meta_data).execute()

        return {"ok": True, "chunks": len(chunks)}
    except Exception as e:
        print(f"Error ingesting team document: {e}")
        raise HTTPException(status_code=500, detail=str(e))
    finally:
        if file_path.exists():
            file_path.unlink()

@router.get("/team/list")
async def list_team_documents(user_id: str = Depends(get_current_user_id)):
    if not user_id:
        raise HTTPException(status_code=401, detail="Unauthorized")

    org_id = await get_user_org(user_id)
    if not org_id:
        return {"documents": []}

    supabase = get_supabase_client()
    try:
        response = supabase.table("team_documents_meta") \
            .select("filename, size_bytes, uploaded_at, uploaded_by") \
            .eq("org_id", org_id) \
            .execute()

        # Deduplicate by filename (in case of multiple uploads)
        unique_docs = {}
        for doc in response.data:
            filename = doc["filename"]
            if filename not in unique_docs:
                unique_docs[filename] = doc

        return {"documents": list(unique_docs.values())}
    except Exception as e:
        print(f"Error listing team documents: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.delete("/team/document/{filename}")
async def delete_team_document(filename: str, user_id: str = Depends(get_current_user_id)):
    if not user_id:
        raise HTTPException(status_code=401, detail="Unauthorized")

    org_id = await get_user_org(user_id)
    if not org_id:
        raise HTTPException(status_code=403, detail="User is not part of an enterprise organization")

    # We should ideally check if user is an admin here, but for now we'll rely on frontend hiding the button

    supabase = get_supabase_client()
    try:
        # We need to query team_documents by JSONB metadata...
        # Fortunately, supabase allows filtering on JSONB fields.
        supabase.table("team_documents") \
            .delete() \
            .eq("org_id", org_id) \
            .eq("metadata->>source_file", filename) \
            .execute()

        supabase.table("team_documents_meta") \
            .delete() \
            .eq("org_id", org_id) \
            .eq("filename", filename) \
            .execute()

        return {"ok": True}
    except Exception as e:
        print(f"Error deleting team document: {e}")
        raise HTTPException(status_code=500, detail=str(e))
