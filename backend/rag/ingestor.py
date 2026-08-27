import pathlib
import shutil

import chromadb
from fastapi import APIRouter, UploadFile, File
from pypdf import PdfReader
from sentence_transformers import SentenceTransformer

from config import config

router = APIRouter()

# Module-level singletons, initialized lazily on first call
_chroma_client = None
_collection = None
_embedder = None


def get_chroma_client():
    global _chroma_client
    if _chroma_client is None:
        _chroma_client = chromadb.PersistentClient(path=config.CHROMA_DIR)
    return _chroma_client


def get_collection():
    global _collection
    if _collection is None:
        _collection = get_chroma_client().get_or_create_collection("user_knowledge")
    return _collection


def get_embedder():
    global _embedder
    if _embedder is None:
        _embedder = SentenceTransformer(config.EMBEDDING_MODEL)
    return _embedder


# Initialize embedder at module import time so download happens at server startup
get_embedder()


def chunk_text(text: str, chunk_size: int = 512, overlap: int = 64) -> list[str]:
    """Splits text into overlapping chunks by character count."""
    if not text:
        return []

    chunks = []
    start = 0
    while start < len(text):
        end = min(start + chunk_size, len(text))
        chunks.append(text[start:end])
        if end == len(text):
            break
        start = end - overlap
    return chunks


def extract_text(filepath: pathlib.Path) -> str:
    """Extracts text from a .pdf, .txt, or .md file."""
    ext = filepath.suffix.lower()

    if ext == ".pdf":
        reader = PdfReader(filepath)
        pages = []
        for page in reader.pages:
            text = page.extract_text()
            if text:
                pages.append(text)
        return "\n".join(pages)
    elif ext in {".txt", ".md"}:
        return filepath.read_text(encoding="utf-8")
    else:
        raise ValueError(f"Unsupported file type: {ext}")


def ingest_file(filepath: pathlib.Path) -> dict:
    """Orchestrates text extraction, chunking, embedding, and storage in ChromaDB."""
    text = extract_text(filepath)
    if not text.strip():
        raise ValueError("Document appears to be empty or unreadable")

    chunks = chunk_text(text)
    print(f"Ingesting {filepath.name}: generated {len(chunks)} chunks.")

    embedder = get_embedder()
    embeddings = embedder.encode(chunks, batch_size=32, show_progress_bar=False).tolist()

    collection = get_collection()

    ids = [f"{filepath.stem}_chunk_{i}" for i in range(len(chunks))]
    metadatas = [
        {
            "source_file": filepath.name,
            "chunk_index": i,
            "file_type": filepath.suffix.lower()
        }
        for i in range(len(chunks))
    ]

    collection.upsert(
        ids=ids,
        embeddings=embeddings,
        documents=chunks,
        metadatas=metadatas
    )

    return {"file": filepath.name, "chunks_ingested": len(chunks)}


@router.post("/upload")
async def upload_document(file: UploadFile = File(...)):
    upload_dir = pathlib.Path(config.UPLOAD_DIR)

    # Sanitize filename to prevent path traversal
    safe_filename = pathlib.Path(file.filename).name
    file_path = upload_dir / safe_filename

    try:
        with open(file_path, "wb") as buffer:
            shutil.copyfileobj(file.file, buffer)

        result = ingest_file(file_path)
        return {"ok": True, "chunks": result["chunks_ingested"]}
    finally:
        if file_path.exists():
            file_path.unlink()


@router.delete("/document/{filename}")
async def delete_document(filename: str):
    collection = get_collection()
    collection.delete(where={"source_file": filename})
    return {"ok": True}


@router.get("/documents")
async def list_documents():
    collection = get_collection()
    result = collection.get(include=["metadatas"])

    documents = set()
    for meta in result.get("metadatas", []):
        if meta and "source_file" in meta:
            documents.add(meta["source_file"])

    return {"documents": list(documents)}
